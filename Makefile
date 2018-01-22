IMAGES=rootfs supermicro-img supermicro-iso

.PHONY: run pkgs zededa-container help $(addsuffix images/,${IMAGES})

all: help

help:
	@echo zenbuild: LinuxKit/moby based Xen images composer
	@echo
	@echo amd64 targets:
	@echo "   'make supermicro.iso' builds a bootable ISO"
	@echo "   'make supermicro.img' builds a bootable raw disk image"


pkgs:
	make -C pkg

run:
	qemu-system-x86_64 --bios ./bios/OVMF.fd -m 4096 -cpu SandyBridge -serial mon:stdio -hda ./supermicro.img \
				-net nic,vlan=0 -net user,id=eth0,vlan=0,net=192.168.1.0/24,dhcpstart=192.168.1.10,hostfwd=tcp::2222-:22 \
				-net nic,vlan=1 -net user,id=eth1,vlan=1,net=192.168.2.0/24,dhcpstart=192.168.2.10

zededa-container/Dockerfile: pkgs parse-pkgs.sh zededa-container/Dockerfile.template
	./parse-pkgs.sh zededa-container/Dockerfile.template > zededa-container/Dockerfile

zededa-container: zededa-container/Dockerfile
	linuxkit pkg build --disable-content-trust zededa-container/

images/%.yml: parse-pkgs.sh images/%.yml.in FORCE
	./parse-pkgs.sh $@.in > $@

supermicro.iso: zededa-container images/supermicro-iso.yml
	./makeiso.sh images/supermicro-iso.yml supermicro.iso

supermicro.img: zededa-container images/supermicro-img.yml
	./makeraw.sh images/supermicro-img.yml supermicro.img

rootfs.img: zededa-container images/rootfs.yml
	./makerootfs.sh images/rootfs.yml image.img


.PHONY: FORCE
FORCE:
