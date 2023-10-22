// Copyright (c) 2017-2023 Zededa, Inc.
// SPDX-License-Identifier: Apache-2.0

package hypervisor

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	zconfig "github.com/lf-edge/eve-api/go/config"
	"github.com/lf-edge/eve/pkg/pillar/agentlog"
	"github.com/lf-edge/eve/pkg/pillar/containerd"
	"github.com/lf-edge/eve/pkg/pillar/types"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

// KVMHypervisorName is a name of kvm hypervisor
const KVMHypervisorName = "kvm"
const minUringKernelTag = uint64((5 << 16) | (4 << 8) | (72 << 0))

// We build device model around PCIe topology according to best practices
//    https://github.com/qemu/qemu/blob/master/docs/pcie.txt
// and
//    https://libvirt.org/pci-hotplug.html
// Thus the only PCI devices plugged directly into the root (pci.0) bus are:
//    00:01.0 cirrus-vga
//    00:02.0 pcie-root-port for QEMU XHCI Host Controller
//    00:03.0 virtio-serial for hvc consoles and serial communications with the domain
//    00:0x.0 pcie-root-port for block or network device #x (where x > 2)
//    00:0y.0 virtio-9p-pci
//
// This makes everything but 9P volumes be separated from root pci bus
// and effectively hang off the bus of its own:
//     01:00.0 QEMU XHCI Host Controller (behind pcie-root-port 00:02.0)
//     xx:00.0 block or network device #x (behind pcie-root-port 00:0x.0)
//
// It would be nice to figure out how to do the same with virtio-9p-pci
// eventually, but for now this is not a high priority.
//
// As discussed in https://edk2.groups.io/g/discuss/topic/windows_2019_vm_fails_to_boot/74465994
// I/O size exceeds the max SCSI I/O limitation(8M) of vhost-scsi in KVM
// we adjust max_sectors option (16384) to run Windows VM with vhost-scsi-pci and avoid errors like
// [ 259.573575] vhost_scsi_calc_sgls: requested sgl_count: 2649 exceeds pre-allocated max_sgls: 2048

const qemuConfTemplate = `# This file is automatically generated by domainmgr
[msg]
  timestamp = "on"

[machine]
  type = "{{.Machine}}"
  dump-guest-core = "off"
{{- if .DomainStatus.CPUs }}
  cpumask = "{{.DomainStatus.CPUs}}"
{{- end -}}
{{- if .DomainConfig.CPUsPinned }}
  cpu-pin = "on"
{{- end -}}
{{- if eq .Machine "virt" }}
  accel = "kvm:tcg"
  gic-version = "host"
{{- end -}}
{{- if ne .Machine "virt" }}
  accel = "kvm"
  vmport = "off"
  kernel-irqchip = "on"
{{- end -}}
{{- if .DomainConfig.BootLoader }}
  firmware = "{{.DomainConfig.BootLoader}}"
{{- end -}}
{{- if .DomainConfig.Kernel }}
  kernel = "{{.DomainConfig.Kernel}}"
{{- end -}}
{{- if .DomainConfig.Ramdisk }}
  initrd = "{{.DomainConfig.Ramdisk}}"
{{- end -}}
{{- if .DomainConfig.DeviceTree }}
  dtb = "{{.DomainConfig.DeviceTree}}"
{{- end -}}
{{- if .DomainConfig.ExtraArgs }}
  append = "{{.DomainConfig.ExtraArgs}}"
{{ end }}
{{if ne .Machine "virt" }}
[global]
  driver = "kvm-pit"
  property = "lost_tick_policy"
  value = "delay"

[global]
  driver = "ICH9-LPC"
  property = "disable_s3"
  value = "1"

[global]
  driver = "ICH9-LPC"
  property = "disable_s4"
  value = "1"

[rtc]
  base = "localtime"
  driftfix = "slew"

[device]
  driver = "intel-iommu"
  caching-mode = "on"
{{ end }}
[realtime]
  mlock = "off"

[chardev "charmonitor"]
  backend = "socket"
  path = "` + kvmStateDir + `{{.DomainConfig.DisplayName}}/qmp"
  server = "on"
  wait = "off"

[mon "monitor"]
  chardev = "charmonitor"
  mode = "control"

[chardev "charlistener"]
  backend = "socket"
  path = "` + kvmStateDir + `{{.DomainConfig.DisplayName}}/listener.qmp"
  server = "on"
  wait = "off"

[mon "listener"]
  chardev = "charlistener"
  mode = "control"

[memory]
  size = "{{.DomainConfig.Memory}}"

[smp-opts]
  cpus = "{{.DomainConfig.VCpus}}"
  sockets = "1"
  cores = "{{.DomainConfig.VCpus}}"
  threads = "1"

[device]
  driver = "virtio-serial"
  addr = "3"

[chardev "charserial0"]
  backend = "socket"
  mux = "on"
  path = "` + kvmStateDir + `{{.DomainConfig.DisplayName}}/cons"
  server = "on"
  wait = "off"
  logfile = "/dev/fd/1"
  logappend = "on"

[device]
  driver = "virtconsole"
  chardev = "charserial0"
  name = "org.lfedge.eve.console.0"

{{if .DomainConfig.IsOCIContainer}}
[chardev "charserial1"]
  backend = "socket"
  mux = "on"
  path = "` + kvmStateDir + `{{.DomainConfig.DisplayName}}/prime-cons"
  server = "on"
  wait = "off"
  logfile = "/dev/fd/1"
  logappend = "on"

[device]
  driver = "virtconsole"
  chardev = "charserial1"
  name = "org.lfedge.eve.console.prime"
{{end}}

{{if .DomainConfig.EnableVnc}}
[vnc "default"]
  vnc = "0.0.0.0:{{if .DomainConfig.VncDisplay}}{{.DomainConfig.VncDisplay}}{{else}}0{{end}}"
  to = "99"
{{- if .DomainConfig.VncPasswd}}
  password = "on"
{{- end -}}
{{end}}
#[device "video0"]
#  driver = "qxl-vga"
#  ram_size = "67108864"
#  vram_size = "67108864"
#  vram64_size_mb = "0"
#  vgamem_mb = "16"
#  max_outputs = "1"
#  bus = "pcie.0"
#  addr = "0x1"
{{ if ne .DomainConfig.GPUConfig "" -}}
{{- if ne .Machine "virt" }}
[device "video0"]
  driver = "VGA"
  vgamem_mb = "16"
  bus = "pcie.0"
  addr = "0x1"
{{else}}
[device "video0"]
  driver = "virtio-gpu-pci"
{{end}}
{{- end}}
[device "pci.2"]
  driver = "pcie-root-port"
  port = "12"
  chassis = "2"
  bus = "pcie.0"
  addr = "0x2"

[device "usb"]
  driver = "qemu-xhci"
  p2 = "15"
  p3 = "15"
  bus = "pci.2"
  addr = "0x0"
{{if ne .Machine "virt" }}
[device "input0"]
  driver = "usb-tablet"
  bus = "usb.0"
  port = "1"
{{else}}
[device "input0"]
  driver = "usb-kbd"
  bus = "usb.0"
  port = "1"

[device "input1"]
  driver = "usb-mouse"
  bus = "usb.0"
  port = "2"
{{end}}`

// multidevs = "remap"
const qemuDiskTemplate = `
{{if eq .Devtype "cdrom"}}
[drive "drive-sata0-{{.DiskID}}"]
  file = "{{.FileLocation}}"
  format = "{{.Format | Fmt}}"
  if = "none"
  media = "cdrom"
  readonly = "on"

[device "sata0-{{.SATAId}}"]
  drive = "drive-sata0-{{.DiskID}}"
{{- if eq .Machine "virt"}}
  driver = "usb-storage"
{{else}}
  driver = "ide-cd"
  bus = "ide.{{.SATAId}}"
{{- end }}
{{else if eq .Devtype "9P"}}
[fsdev "fsdev{{.DiskID}}"]
  fsdriver = "local"
  security_model = "none"
  path = "{{.FileLocation}}"

[device "fs{{.DiskID}}"]
  driver = "virtio-9p-pci"
  fsdev = "fsdev{{.DiskID}}"
  mount_tag = "share_dir"
  addr = "{{printf "0x%x" .PCIId}}"
{{else}}
[device "pci.{{.PCIId}}"]
  driver = "pcie-root-port"
  port = "1{{.PCIId}}"
  chassis = "{{.PCIId}}"
  bus = "pcie.0"
  addr = "{{printf "0x%x" .PCIId}}"
{{if eq .WWN ""}}
[drive "drive-virtio-disk{{.DiskID}}"]
  file = "{{.FileLocation}}"
  format = "{{.Format | Fmt}}"
  aio = "{{.AioType}}"
  cache = "writeback"
  if = "none"
{{if .ReadOnly}}  readonly = "on"{{end}}
{{- if eq .Devtype "legacy"}}
[device "ahci.{{.PCIId}}"]
  bus = "pci.{{.PCIId}}"
  driver = "ahci"

[device "ahci-disk{{.DiskID}}"]
  driver = "ide-hd"
  bus = "ahci.{{.PCIId}}.0"
{{- else}}
[device "virtio-disk{{.DiskID}}"]
  driver = "virtio-blk-pci"
  scsi = "off"
  bus = "pci.{{.PCIId}}"
  addr = "0x0"
{{- end}}
  drive = "drive-virtio-disk{{.DiskID}}"
{{- else}}
[device "vhost-disk{{.DiskID}}"]
  driver = "vhost-scsi-pci"
  max_sectors = "16384"
  wwpn = "{{.WWN}}"
  bus = "pci.{{.PCIId}}"
  addr = "0x0"
  num_queues = "{{.NumQueues}}"
{{- end}}
{{end}}`

const qemuNetTemplate = `
[device "pci.{{.PCIId}}"]
  driver = "pcie-root-port"
  port = "1{{.PCIId}}"
  chassis = "{{.PCIId}}"
  bus = "pcie.0"
  multifunction = "on"
  addr = "{{printf "0x%x" .PCIId}}"

[netdev "hostnet{{.NetID}}"]
  type = "tap"
  ifname = "{{.Vif}}"
  br = "{{.Bridge}}"
  script = "/etc/xen/scripts/qemu-ifup"
  downscript = "no"

[device "net{{.NetID}}"]
  driver = "{{.Driver}}"
  netdev = "hostnet{{.NetID}}"
  mac = "{{.Mac}}"
  bus = "pci.{{.PCIId}}"
  addr = "0x0"
`

const qemuPciPassthruTemplate = `
[device "pci.{{.PCIId}}"]
  driver = "pcie-root-port"
  port = "1{{.PCIId}}"
  chassis = "{{.PCIId}}"
  bus = "pcie.0"
  multifunction = "on"
  addr = "{{printf "0x%x" .PCIId}}"

[device]
  driver = "vfio-pci"
  host = "{{.PciShortAddr}}"
  bus = "pci.{{.PCIId}}"
  addr = "0x0"
{{- if .Xvga }}
  x-vga = "on"
{{- end -}}
{{- if .Xopregion }}
  x-igd-opregion = "on"
{{- end -}}
`
const qemuSerialTemplate = `
[chardev "charserial-usr{{.ID}}"]
{{- if eq .Machine "virt"}}
  backend = "serial"
{{- else}}
  backend = "tty"
{{- end}}
  path = "{{.SerialPortName}}"

[device "serial-usr{{.ID}}"]
{{- if eq .Machine "virt"}}
  driver = "pci-serial"
{{- else}}
  driver = "isa-serial"
{{- end}}
  chardev = "charserial-usr{{.ID}}"
`

const qemuUsbHostTemplate = `
[device]
  driver = "usb-host"
  hostbus = "{{.UsbBus}}"
  hostaddr = "{{.UsbDevAddr}}"
`

const kvmStateDir = "/run/hypervisor/kvm/"
const sysfsVfioPciBind = "/sys/bus/pci/drivers/vfio-pci/bind"
const sysfsPciDriversProbe = "/sys/bus/pci/drivers_probe"
const vfioDriverPath = "/sys/bus/pci/drivers/vfio-pci"

// KVM domains map 1-1 to anchor device model UNIX processes (qemu or firecracker)
// For every anchor process we maintain the following entry points in the
// /run/hypervisor/kvm/DOMAIN_NAME:
//
//	 pid - contains PID of the anchor process
//	 qmp - UNIX domain socket that allows us to talk to anchor process
//	cons - symlink to /dev/pts/X that allows us to talk to the serial console of the domain
//
// In addition to that, we also maintain DOMAIN_NAME -> PID mapping in kvmContext, so we don't
// have to look things up in the filesystem all the time (this also allows us to filter domains
// that may be created by others)
type kvmContext struct {
	ctrdContext
	// for now the following is statically configured and can not be changed per domain
	devicemodel  string
	dmExec       string
	dmArgs       []string
	dmCPUArgs    []string
	dmFmlCPUArgs []string
	capabilities *types.Capabilities
}

func newKvm() Hypervisor {
	ctrdCtx, err := initContainerd()
	if err != nil {
		logrus.Fatalf("couldn't initialize containerd (this should not happen): %v. Exiting.", err)
		return nil // it really never returns on account of above
	}
	// later on we may want to pass device model machine type in DomainConfig directly;
	// for now -- lets just pick a static device model based on the host architecture
	// "-cpu host",
	// -cpu IvyBridge-IBRS,ss=on,vmx=on,movbe=on,hypervisor=on,arat=on,tsc_adjust=on,mpx=on,rdseed=on,smap=on,clflushopt=on,sha-ni=on,umip=on,md-clear=on,arch-capabilities=on,xsaveopt=on,xsavec=on,xgetbv1=on,xsaves=on,pdpe1gb=on,3dnowprefetch=on,avx=off,f16c=off,hv_time,hv_relaxed,hv_vapic,hv_spinlocks=0x1fff
	switch runtime.GOARCH {
	case "arm64":
		return kvmContext{
			ctrdContext:  *ctrdCtx,
			devicemodel:  "virt",
			dmExec:       "/usr/lib/xen/bin/qemu-system-aarch64",
			dmArgs:       []string{"-display", "none", "-S", "-no-user-config", "-nodefaults", "-no-shutdown", "-overcommit", "mem-lock=on", "-overcommit", "cpu-pm=on", "-serial", "chardev:charserial0"},
			dmCPUArgs:    []string{"-cpu", "host"},
			dmFmlCPUArgs: []string{"-cpu", "host"},
		}
	case "amd64":
		return kvmContext{
			//nolint:godox // FIXME: Removing "-overcommit", "mem-lock=on", "-overcommit" for now, revisit it later as part of resource partitioning
			ctrdContext:  *ctrdCtx,
			devicemodel:  "pc-q35-3.1",
			dmExec:       "/usr/lib/xen/bin/qemu-system-x86_64",
			dmArgs:       []string{"-display", "none", "-S", "-no-user-config", "-nodefaults", "-no-shutdown", "-serial", "chardev:charserial0", "-no-hpet"},
			dmCPUArgs:    []string{"-cpu", "host"},
			dmFmlCPUArgs: []string{"-cpu", "host,hv_time,hv_relaxed,hv_vendor_id=eveitis,hypervisor=off,kvm=off"},
		}
	}
	return nil
}

func (ctx kvmContext) GetCapabilities() (*types.Capabilities, error) {
	if ctx.capabilities != nil {
		return ctx.capabilities, nil
	}
	vtd, err := ctx.checkIOVirtualisation()
	if err != nil {
		return nil, fmt.Errorf("fail in check IOVirtualization: %v", err)
	}
	ctx.capabilities = &types.Capabilities{
		HWAssistedVirtualization: true,
		IOVirtualization:         vtd,
		CPUPinning:               true,
		UseVHost:                 true,
	}
	return ctx.capabilities, nil
}

// CountMemOverhead - returns the memory overhead estimation for a domain.
func (ctx kvmContext) CountMemOverhead(domainName string, domainUUID uuid.UUID, domainRAMSize int64, vmmMaxMem int64,
	domainMaxCpus int64, domainVCpus int64, domainIoAdapterList []types.IoAdapter, aa *types.AssignableAdapters,
	globalConfig *types.ConfigItemValueMap) (uint64, error) {
	result, err := vmmOverhead(domainName, domainUUID, domainRAMSize, vmmMaxMem, domainMaxCpus, domainVCpus, domainIoAdapterList, aa, globalConfig)
	return uint64(result), err
}

func (ctx kvmContext) checkIOVirtualisation() (bool, error) {
	f, err := os.Open("/sys/kernel/iommu_groups")
	if err == nil {
		files, err := f.Readdirnames(0)
		if err != nil {
			return false, err
		}
		if len(files) != 0 {
			return true, nil
		}
	}
	return false, err
}

func (ctx kvmContext) Name() string {
	return KVMHypervisorName
}

func (ctx kvmContext) Task(status *types.DomainStatus) types.Task {
	if status.VirtualizationMode == types.NOHYPER {
		return ctx.ctrdContext
	}
	return ctx
}

func estimatedVMMOverhead(domainName string, aa *types.AssignableAdapters, domainAdapterList []types.IoAdapter,
	domainUUID uuid.UUID, domainRAMSize int64, domainMaxCpus int64, domainVcpus int64) (int64, error) {
	var overhead int64

	mmioOverhead, err := mmioVMMOverhead(domainName, aa, domainAdapterList, domainUUID)

	if err != nil {
		return 0, logError("mmioVMMOverhead() failed for domain %s: %v",
			domainName, err)
	}
	overhead = undefinedVMMOverhead() + ramVMMOverhead(domainRAMSize) +
		qemuVMMOverhead() + cpuVMMOverhead(domainMaxCpus, domainVcpus) + mmioOverhead

	return overhead, nil
}

func ramVMMOverhead(ramMemory int64) int64 {
	// 0.224% of the total RAM allocated for VM in bytes
	// this formula is precise and well explained in the following QEMU issue:
	// https://gitlab.com/qemu-project/qemu/-/issues/1003
	// This is a best case scenario because it assumes that all PTEs are allocated
	// sequentially. In reality, there will be some fragmentation and the overhead
	// for now 2.5% (~10x) is a good approximation until we have a better way to
	// predict the memory usage of the VM.
	return ramMemory * 1024 * 25 / 1000
}

// overhead for qemu binaries and libraries
func qemuVMMOverhead() int64 {
	return 20 << 20 // Mb in bytes
}

// overhead for VMM memory mapped IO
// it fluctuates between 0.66 and 0.81 % of MMIO total size
// for all mapped devices. Set it to 1% to be on the safe side
// this can be a pretty big number for GPUs with very big
// aperture size (e.g. 64G for NVIDIA A40)
func mmioVMMOverhead(domainName string, aa *types.AssignableAdapters, domainAdapterList []types.IoAdapter,
	domainUUID uuid.UUID) (int64, error) {
	var pciAssignments []pciDevice
	var mmioSize uint64

	for _, adapter := range domainAdapterList {
		logrus.Debugf("processing adapter %d %s\n", adapter.Type, adapter.Name)
		aaList := aa.LookupIoBundleAny(adapter.Name)
		// We reserved it in handleCreate so nobody could have stolen it
		if len(aaList) == 0 {
			return 0, logError("IoBundle disappeared %d %s for %s\n",
				adapter.Type, adapter.Name, domainName)
		}
		for _, ib := range aaList {
			if ib == nil {
				continue
			}
			if ib.UsedByUUID != domainUUID {
				return 0, logError("IoBundle not ours %s: %d %s for %s\n",
					ib.UsedByUUID, adapter.Type, adapter.Name,
					domainName)
			}
			if ib.PciLong != "" && ib.UsbAddr == "" {
				logrus.Infof("Adding PCI device <%s>\n", ib.PciLong)
				tap := pciDevice{pciLong: ib.PciLong, ioType: ib.Type}
				pciAssignments = addNoDuplicatePCI(pciAssignments, tap)
			}
		}
	}

	for _, dev := range pciAssignments {
		logrus.Infof("PCI device %s %d\n", dev.pciLong, dev.ioType)
		// read the size of the PCI device aperture. Only GPU/VGA devices for now
		if dev.ioType != types.IoOther && dev.ioType != types.IoHDMI {
			continue
		}
		// skip bridges
		isBridge, err := dev.isBridge()
		if err != nil {
			// do not treat as fatal error
			logrus.Warnf("Can't read PCI device class, treat as bridge %s: %v\n",
				dev.pciLong, err)
			isBridge = true
		}

		if isBridge {
			logrus.Infof("Skipping bridge %s\n", dev.pciLong)
			continue
		}

		// read all resources of the PCI device
		resources, err := dev.readResources()
		if err != nil {
			return 0, logError("Can't read PCI device resources %s: %v\n",
				dev.pciLong, err)
		}

		// calculate the size of the MMIO region
		for _, res := range resources {
			if res.valid() && res.isMem() {
				mmioSize += res.size()
			}
		}
	}

	// 1% of the total MMIO size in bytes
	mmioOverhead := int64(mmioSize) / 100

	logrus.Infof("MMIO size: %d / overhead: %d for %s", mmioSize, mmioOverhead, domainName)

	return int64(mmioOverhead), nil
}

// each vCPU requires about 3MB of memory
func cpuVMMOverhead(maxCpus int64, vcpus int64) int64 {
	cpus := maxCpus
	if cpus == 0 {
		cpus = vcpus
	}
	return cpus * (3 << 20) // Mb in bytes
}

// memory allocated by QEMU for its own purposes.
// statistical analysis did not revile any correlation between
// VM configuration (devices, nr of vcpus, etc) and this number
// however the size of disk space affects it. Probably some internal
// QEMU caches are allocated based on the size of the disk image.
// it requires more investigation.
func undefinedVMMOverhead() int64 {
	return 350 << 20 // Mb in bytes
}

func vmmOverhead(domainName string, domainUUID uuid.UUID, domainRAMSize int64, vmmMaxMem int64, domainMaxCpus int64, domainVCpus int64, domainIoAdapterList []types.IoAdapter, aa *types.AssignableAdapters, globalConfig *types.ConfigItemValueMap) (int64, error) {
	var overhead int64

	// Fetch VMM max memory setting (aka vmm overhead)
	overhead = vmmMaxMem << 10

	// Global node setting has a higher priority
	if globalConfig != nil {
		VmmOverheadOverrideCfgItem, ok := globalConfig.GlobalSettings[types.VmmMemoryLimitInMiB]
		if !ok {
			return 0, logError("Missing key %s", string(types.VmmMemoryLimitInMiB))
		}
		if VmmOverheadOverrideCfgItem.IntValue > 0 {
			overhead = int64(VmmOverheadOverrideCfgItem.IntValue) << 20
		}
	}

	if overhead == 0 {
		overhead, err := estimatedVMMOverhead(domainName, aa, domainIoAdapterList, domainUUID, domainRAMSize, domainMaxCpus, domainVCpus)
		if err != nil {
			return 0, logError("estimatedVMMOverhead() failed for domain %s: %v",
				domainName, err)
		}
		return overhead, nil
	}

	return overhead, nil
}

func (ctx kvmContext) Setup(status types.DomainStatus, config types.DomainConfig,
	aa *types.AssignableAdapters, globalConfig *types.ConfigItemValueMap, file *os.File) error {

	diskStatusList := status.DiskStatusList
	domainName := status.DomainName
	domainUUID := status.UUIDandVersion.UUID
	// first lets build the domain config
	if err := ctx.CreateDomConfig(domainName, config, status, diskStatusList, aa, file); err != nil {
		return logError("failed to build domain config: %v", err)
	}

	dmArgs := ctx.dmArgs
	if config.VirtualizationMode == types.FML {
		dmArgs = append(dmArgs, ctx.dmFmlCPUArgs...)
	} else {
		dmArgs = append(dmArgs, ctx.dmCPUArgs...)
	}

	if config.MetaDataType == types.MetaDataOpenStack {
		// we need to set product_name to support cloud-init
		dmArgs = append(dmArgs, "-smbios", "type=1,product=OpenStack Compute")
	}

	os.MkdirAll(kvmStateDir+domainName, 0777)

	args := []string{ctx.dmExec}
	args = append(args, dmArgs...)
	args = append(args, "-name", domainName,
		"-uuid", domainUUID.String(),
		"-readconfig", file.Name(),
		"-pidfile", kvmStateDir+domainName+"/pid")

	spec, err := ctx.setupSpec(&status, &config, status.OCIConfigDir)
	if err != nil {
		return logError("failed to load OCI spec for domain %s: %v", status.DomainName, err)
	}
	if err = spec.AddLoader("/containers/services/xen-tools"); err != nil {
		return logError("failed to add kvm hypervisor loader to domain %s: %v", status.DomainName, err)
	}
	overhead, err := vmmOverhead(domainName, domainUUID, int64(config.Memory), int64(config.VMMMaxMem), int64(config.MaxCpus), int64(config.VCpus), config.IoAdapterList, aa, globalConfig)
	if err != nil {
		return logError("vmmOverhead() failed for domain %s: %v",
			status.DomainName, err)
	}
	logrus.Debugf("Qemu overhead for domain %s is %d bytes", status.DomainName, overhead)
	spec.AdjustMemLimit(config, overhead)
	spec.Get().Process.Args = args
	logrus.Infof("Hypervisor args: %v", args)

	if err := spec.CreateContainer(true); err != nil {
		return logError("Failed to create container for task %s from %v: %v", status.DomainName, config, err)
	}

	return nil
}

func (ctx kvmContext) CreateDomConfig(domainName string, config types.DomainConfig, status types.DomainStatus,
	diskStatusList []types.DiskStatus, aa *types.AssignableAdapters, file *os.File) error {
	tmplCtx := struct {
		Machine string
		types.DomainConfig
		types.DomainStatus
	}{ctx.devicemodel, config, status}
	tmplCtx.DomainConfig.Memory = (config.Memory + 1023) / 1024
	tmplCtx.DomainConfig.DisplayName = domainName

	// render global device model settings
	t, _ := template.New("qemu").Parse(qemuConfTemplate)
	if err := t.Execute(file, tmplCtx); err != nil {
		return logError("can't write to config file %s (%v)", file.Name(), err)
	}

	// render disk device model settings
	diskContext := struct {
		Machine                          string
		PCIId, DiskID, SATAId, NumQueues int
		AioType                          string
		types.DiskStatus
	}{Machine: ctx.devicemodel, PCIId: 4, DiskID: 0, SATAId: 0, AioType: "io_uring", NumQueues: config.VCpus}

	t, _ = template.New("qemuDisk").
		Funcs(template.FuncMap{"Fmt": func(f zconfig.Format) string { return strings.ToLower(f.String()) }}).
		Parse(qemuDiskTemplate)
	for _, ds := range diskStatusList {
		if ds.Devtype == "" {
			continue
		}
		if ds.Devtype == "AppCustom" {
			// This is application custom data. It is forwarded to the VM
			// differently - as a download url in zedrouter
			continue
		}
		diskContext.DiskStatus = ds
		if err := t.Execute(file, diskContext); err != nil {
			return logError("can't write to config file %s (%v)", file.Name(), err)
		}
		if diskContext.Devtype == "cdrom" {
			diskContext.SATAId = diskContext.SATAId + 1
		} else {
			diskContext.PCIId = diskContext.PCIId + 1
		}
		diskContext.DiskID = diskContext.DiskID + 1
	}

	// render network device model settings
	netContext := struct {
		PCIId, NetID     int
		Driver           string
		Mac, Bridge, Vif string
	}{PCIId: diskContext.PCIId, NetID: 0}
	t, _ = template.New("qemuNet").Parse(qemuNetTemplate)
	for _, net := range config.VifList {
		netContext.Mac = net.Mac.String()
		netContext.Bridge = net.Bridge
		netContext.Vif = net.Vif
		if config.VirtualizationMode == types.LEGACY {
			netContext.Driver = "e1000"
		} else {
			netContext.Driver = "virtio-net-pci"
		}
		if err := t.Execute(file, netContext); err != nil {
			return logError("can't write to config file %s (%v)", file.Name(), err)
		}
		netContext.PCIId = netContext.PCIId + 1
		netContext.NetID = netContext.NetID + 1
	}

	// Gather all PCI assignments into a single line
	var pciAssignments []pciDevice
	// Gather all USB assignments into a single line
	var usbAssignments []string
	// Gather all serial assignments into a single line
	var serialAssignments []string

	for _, adapter := range config.IoAdapterList {
		logrus.Debugf("processing adapter %d %s\n", adapter.Type, adapter.Name)
		list := aa.LookupIoBundleAny(adapter.Name)
		// We reserved it in handleCreate so nobody could have stolen it
		if len(list) == 0 {
			logrus.Fatalf("IoBundle disappeared %d %s for %s\n",
				adapter.Type, adapter.Name, domainName)
		}
		for _, ib := range list {
			if ib == nil {
				continue
			}
			if ib.UsedByUUID != config.UUIDandVersion.UUID {
				logrus.Fatalf("IoBundle not ours %s: %d %s for %s\n",
					ib.UsedByUUID, adapter.Type, adapter.Name,
					domainName)
			}
			if ib.PciLong != "" {
				logrus.Infof("Adding PCI device <%v>\n", ib.PciLong)
				tap := pciDevice{pciLong: ib.PciLong, ioType: ib.Type}
				pciAssignments = addNoDuplicatePCI(pciAssignments, tap)
			}
			if ib.Serial != "" {
				logrus.Infof("Adding serial <%s>\n", ib.Serial)
				serialAssignments = addNoDuplicate(serialAssignments, ib.Serial)
			}
			if ib.UsbAddr != "" {
				logrus.Infof("Adding USB host device <%s>\n", ib.UsbAddr)
				usbAssignments = addNoDuplicate(usbAssignments, ib.UsbAddr)
			}
		}
	}
	if len(pciAssignments) != 0 {
		pciPTContext := struct {
			PCIId        int
			PciShortAddr string
			Xvga         bool
			Xopregion    bool
		}{PCIId: netContext.PCIId, PciShortAddr: "", Xvga: false, Xopregion: false}

		t, _ = template.New("qemuPciPT").Parse(qemuPciPassthruTemplate)
		for _, pa := range pciAssignments {
			short := types.PCILongToShort(pa.pciLong)
			pciPTContext.Xvga = pa.isVGA()

			if vendor, err := pa.vid(); err == nil {
				// check for Intel vendor
				if vendor == "0x8086" {
					if pciPTContext.Xvga {
						// we set opregion for Intel vga
						// https://github.com/qemu/qemu/blob/stable-5.0/docs/igd-assign.txt#L91-L96
						pciPTContext.Xopregion = true
					}
				}
			}

			pciPTContext.PciShortAddr = short
			if err := t.Execute(file, pciPTContext); err != nil {
				return logError("can't write PCI Passthrough to config file %s (%v)", file.Name(), err)
			}
			pciPTContext.Xvga = false
			pciPTContext.Xopregion = false
			pciPTContext.PCIId = pciPTContext.PCIId + 1
		}
	}
	if len(serialAssignments) != 0 {
		serialPortContext := struct {
			Machine        string
			SerialPortName string
			ID             int
		}{Machine: ctx.devicemodel, SerialPortName: "", ID: 0}

		t, _ = template.New("qemuSerial").Parse(qemuSerialTemplate)
		for id, serial := range serialAssignments {
			serialPortContext.SerialPortName = serial
			fmt.Printf("id for serial is %d\n", id)
			serialPortContext.ID = id
			if err := t.Execute(file, serialPortContext); err != nil {
				return logError("can't write serial assignment to config file %s (%v)", file.Name(), err)
			}
		}
	}
	if len(usbAssignments) != 0 {
		usbHostContext := struct {
			UsbBus     string
			UsbDevAddr string
			// Ports are dot-separated
		}{UsbBus: "", UsbDevAddr: ""}

		t, _ = template.New("qemuUsbHost").Parse(qemuUsbHostTemplate)
		for _, usbaddr := range usbAssignments {
			bus, port := usbBusPort(usbaddr)
			usbHostContext.UsbBus = bus
			usbHostContext.UsbDevAddr = port
			if err := t.Execute(file, usbHostContext); err != nil {
				return logError("can't write USB host device assignment to config file %s (%v)", file.Name(), err)
			}
		}
	}

	return nil
}

func waitForQmp(domainName string, available bool) error {
	maxDelay := time.Second * 10
	delay := time.Second
	var waited time.Duration
	var err error
	for {
		logrus.Infof("waitForQmp for %s %t: waiting for %v", domainName, available, delay)
		if delay != 0 {
			time.Sleep(delay)
			waited += delay
		}
		if _, err := getQemuStatus(getQmpExecutorSocket(domainName)); available == (err == nil) {
			logrus.Infof("waitForQmp for %s %t done", domainName, available)
			return nil
		}
		if waited > maxDelay {
			// Give up
			logrus.Warnf("waitForQmp for %s %t: giving up", domainName, available)
			if available {
				return logError("Qmp not found: error %v", err)
			}
			return logError("Qmp still available")
		}
		delay = 2 * delay
		if delay > time.Minute {
			delay = time.Minute
		}
	}
}

func (ctx kvmContext) Start(domainName string) error {
	logrus.Infof("starting KVM domain %s", domainName)
	if err := ctx.ctrdContext.Start(domainName); err != nil {
		logrus.Errorf("couldn't start task for domain %s: %v", domainName, err)
		return err
	}
	logrus.Infof("done launching qemu device model")
	if err := waitForQmp(domainName, true); err != nil {
		logrus.Errorf("Error waiting for Qmp for domain %s: %v", domainName, err)
		return err
	}
	logrus.Infof("done launching qemu device model")

	qmpFile := getQmpExecutorSocket(domainName)

	logrus.Debugf("starting qmpEventHandler")
	logrus.Infof("Creating %s at %s", "qmpEventHandler", agentlog.GetMyStack())
	go qmpEventHandler(getQmpListenerSocket(domainName), getQmpExecutorSocket(domainName))

	annotations, err := ctx.ctrdContext.Annotations(domainName)
	if err != nil {
		logrus.Warnf("Error in get annotations for domain %s: %v", domainName, err)
		return err
	}

	if vncPassword, ok := annotations[containerd.EVEOCIVNCPasswordLabel]; ok && vncPassword != "" {
		if err := execVNCPassword(qmpFile, vncPassword); err != nil {
			return logError("failed to set VNC password %v", err)
		}
	}

	if err := execContinue(qmpFile); err != nil {
		return logError("failed to start domain that is stopped %v", err)
	}

	if status, err := getQemuStatus(qmpFile); err != nil || status != "running" {
		return logError("domain status is not running but %s after cont command returned %v", status, err)
	}
	return nil
}

func (ctx kvmContext) Stop(domainName string, _ bool) error {
	if err := execShutdown(getQmpExecutorSocket(domainName)); err != nil {
		return logError("Stop: failed to execute shutdown command %v", err)
	}
	return nil
}

func (ctx kvmContext) Delete(domainName string) (result error) {
	//Sending a stop signal to then domain before quitting. This is done to freeze the domain before quitting it.
	execStop(getQmpExecutorSocket(domainName))
	if err := execQuit(getQmpExecutorSocket(domainName)); err != nil {
		return logError("failed to execute quit command %v", err)
	}
	// we may want to wait a little bit here and actually kill qemu process if it gets wedged
	if err := os.RemoveAll(kvmStateDir + domainName); err != nil {
		return logError("failed to clean up domain state directory %s (%v)", domainName, err)
	}

	return nil
}

func (ctx kvmContext) Info(domainName string) (int, types.SwState, error) {
	// first we ask for the task status
	effectiveDomainID, effectiveDomainState, err := ctx.ctrdContext.Info(domainName)
	if err != nil || effectiveDomainState != types.RUNNING {
		return effectiveDomainID, effectiveDomainState, err
	}

	// if task us alive, we augment task status with finer grained details from qemu
	// lets parse the status according to https://github.com/qemu/qemu/blob/master/qapi/run-state.json#L8
	stateMap := map[string]types.SwState{
		"finish-migrate": types.PAUSED,
		"inmigrate":      types.PAUSING,
		"paused":         types.PAUSED,
		"postmigrate":    types.PAUSED,
		"prelaunch":      types.PAUSED,
		"restore-vm":     types.PAUSED,
		"running":        types.RUNNING,
		"save-vm":        types.PAUSED,
		"shutdown":       types.HALTING,
		"suspended":      types.PAUSED,
		"watchdog":       types.PAUSING,
		"colo":           types.PAUSED,
		"preconfig":      types.PAUSED,
	}
	res, err := getQemuStatus(getQmpExecutorSocket(domainName))
	if err != nil {
		return effectiveDomainID, types.BROKEN, logError("couldn't retrieve status for domain %s: %v", domainName, err)
	}

	if effectiveDomainState, matched := stateMap[res]; !matched {
		return effectiveDomainID, types.BROKEN, logError("domain %s reported to be in unexpected state %s", domainName, res)
	} else {
		return effectiveDomainID, effectiveDomainState, nil
	}
}

func (ctx kvmContext) Cleanup(domainName string) error {
	if err := ctx.ctrdContext.Cleanup(domainName); err != nil {
		return fmt.Errorf("couldn't cleanup task %s: %v", domainName, err)
	}
	if err := waitForQmp(domainName, false); err != nil {
		return fmt.Errorf("error waiting for Qmp absent for domain %s: %v", domainName, err)
	}

	return nil
}

func (ctx kvmContext) PCIReserve(long string) error {
	logrus.Infof("PCIReserve long addr is %s", long)

	overrideFile := filepath.Join(sysfsPciDevices, long, "driver_override")
	driverPath := filepath.Join(sysfsPciDevices, long, "driver")
	unbindFile := filepath.Join(driverPath, "unbind")

	//Check if already bound to vfio-pci
	driverPathInfo, driverPathErr := os.Stat(driverPath)
	vfioDriverPathInfo, vfioDriverPathErr := os.Stat(vfioDriverPath)
	if driverPathErr == nil && vfioDriverPathErr == nil &&
		os.SameFile(driverPathInfo, vfioDriverPathInfo) {
		logrus.Infof("Driver for %s is already bound to vfio-pci, skipping unbind", long)
		return nil
	}

	//map vfio-pci as the driver_override for the device
	if err := os.WriteFile(overrideFile, []byte("vfio-pci"), 0644); err != nil {
		return logError("driver_override failure for PCI device %s: %v",
			long, err)
	}

	//Unbind the current driver, whatever it is, if there is one
	if _, err := os.Stat(unbindFile); err == nil {
		if err := os.WriteFile(unbindFile, []byte(long), 0644); err != nil {
			return logError("unbind failure for PCI device %s: %v",
				long, err)
		}
	}

	if err := os.WriteFile(sysfsPciDriversProbe, []byte(long), 0644); err != nil {
		return logError("drivers_probe failure for PCI device %s: %v",
			long, err)
	}

	return nil
}

func (ctx kvmContext) PCIRelease(long string) error {
	logrus.Infof("PCIRelease long addr is %s", long)

	overrideFile := filepath.Join(sysfsPciDevices, long, "driver_override")
	unbindFile := filepath.Join(sysfsPciDevices, long, "driver/unbind")

	//Write Empty string, to clear driver_override for the device
	if err := os.WriteFile(overrideFile, []byte("\n"), 0644); err != nil {
		logrus.Fatalf("driver_override failure for PCI device %s: %v",
			long, err)
	}

	//Unbind vfio-pci, if unbind file is present
	if _, err := os.Stat(unbindFile); err == nil {
		if err := os.WriteFile(unbindFile, []byte(long), 0644); err != nil {
			logrus.Fatalf("unbind failure for PCI device %s: %v",
				long, err)
		}
	}

	//Write PCI DDDD:BB:DD.FF to /sys/bus/pci/drivers_probe,
	//as a best-effort to bring back original driver
	if err := os.WriteFile(sysfsPciDriversProbe, []byte(long), 0644); err != nil {
		logrus.Fatalf("drivers_probe failure for PCI device %s: %v",
			long, err)
	}

	return nil
}

func (ctx kvmContext) PCISameController(id1 string, id2 string) bool {
	tag1, err := types.PCIGetIOMMUGroup(id1)
	if err != nil {
		return types.PCISameController(id1, id2)
	}

	tag2, err := types.PCIGetIOMMUGroup(id2)
	if err != nil {
		return types.PCISameController(id1, id2)
	}

	return tag1 == tag2
}

func usbBusPort(USBAddr string) (string, string) {
	ids := strings.SplitN(USBAddr, ":", 2)
	if len(ids) == 2 {
		return ids[0], ids[1]
	}
	return "", ""
}

func getQmpExecutorSocket(domainName string) string {
	return filepath.Join(kvmStateDir, domainName, "qmp")
}

func getQmpListenerSocket(domainName string) string {
	return filepath.Join(kvmStateDir, domainName, "listener.qmp")
}
