# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: profile/local_profile.proto

from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from info import info_pb2 as info_dot_info__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='profile/local_profile.proto',
  package='org.lfedge.eve.profile',
  syntax='proto3',
  serialized_options=b'\n\026org.lfedge.eve.profileZ%github.com/lf-edge/eve/api/go/profile',
  create_key=_descriptor._internal_create_key,
  serialized_pb=b'\n\x1bprofile/local_profile.proto\x12\x16org.lfedge.eve.profile\x1a\x0finfo/info.proto\";\n\x0cLocalProfile\x12\x15\n\rlocal_profile\x18\x01 \x01(\t\x12\x14\n\x0cserver_token\x18\x02 \x01(\t\"{\n\x0bRadioStatus\x12\x15\n\rradio_silence\x18\x01 \x01(\x08\x12\x14\n\x0c\x63onfig_error\x18\x02 \x01(\t\x12?\n\x0f\x63\x65llular_status\x18\x03 \x03(\x0b\x32&.org.lfedge.eve.profile.CellularStatus\"\xfc\x01\n\x0e\x43\x65llularStatus\x12\x14\n\x0clogicallabel\x18\x01 \x01(\t\x12\x38\n\x06module\x18\x02 \x01(\x0b\x32(.org.lfedge.eve.info.ZCellularModuleInfo\x12\x34\n\tsim_cards\x18\x03 \x03(\x0b\x32!.org.lfedge.eve.info.ZSimcardInfo\x12\x39\n\tproviders\x18\x04 \x03(\x0b\x32&.org.lfedge.eve.info.ZCellularProvider\x12\x14\n\x0c\x63onfig_error\x18\n \x01(\t\x12\x13\n\x0bprobe_error\x18\x0b \x01(\t\":\n\x0bRadioConfig\x12\x14\n\x0cserver_token\x18\x01 \x01(\t\x12\x15\n\rradio_silence\x18\x02 \x01(\x08\x42?\n\x16org.lfedge.eve.profileZ%github.com/lf-edge/eve/api/go/profileb\x06proto3'
  ,
  dependencies=[info_dot_info__pb2.DESCRIPTOR,])




_LOCALPROFILE = _descriptor.Descriptor(
  name='LocalProfile',
  full_name='org.lfedge.eve.profile.LocalProfile',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='local_profile', full_name='org.lfedge.eve.profile.LocalProfile.local_profile', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='server_token', full_name='org.lfedge.eve.profile.LocalProfile.server_token', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=72,
  serialized_end=131,
)


_RADIOSTATUS = _descriptor.Descriptor(
  name='RadioStatus',
  full_name='org.lfedge.eve.profile.RadioStatus',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='radio_silence', full_name='org.lfedge.eve.profile.RadioStatus.radio_silence', index=0,
      number=1, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='config_error', full_name='org.lfedge.eve.profile.RadioStatus.config_error', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='cellular_status', full_name='org.lfedge.eve.profile.RadioStatus.cellular_status', index=2,
      number=3, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=133,
  serialized_end=256,
)


_CELLULARSTATUS = _descriptor.Descriptor(
  name='CellularStatus',
  full_name='org.lfedge.eve.profile.CellularStatus',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='logicallabel', full_name='org.lfedge.eve.profile.CellularStatus.logicallabel', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='module', full_name='org.lfedge.eve.profile.CellularStatus.module', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='sim_cards', full_name='org.lfedge.eve.profile.CellularStatus.sim_cards', index=2,
      number=3, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='providers', full_name='org.lfedge.eve.profile.CellularStatus.providers', index=3,
      number=4, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='config_error', full_name='org.lfedge.eve.profile.CellularStatus.config_error', index=4,
      number=10, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='probe_error', full_name='org.lfedge.eve.profile.CellularStatus.probe_error', index=5,
      number=11, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=259,
  serialized_end=511,
)


_RADIOCONFIG = _descriptor.Descriptor(
  name='RadioConfig',
  full_name='org.lfedge.eve.profile.RadioConfig',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='server_token', full_name='org.lfedge.eve.profile.RadioConfig.server_token', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='radio_silence', full_name='org.lfedge.eve.profile.RadioConfig.radio_silence', index=1,
      number=2, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=513,
  serialized_end=571,
)

_RADIOSTATUS.fields_by_name['cellular_status'].message_type = _CELLULARSTATUS
_CELLULARSTATUS.fields_by_name['module'].message_type = info_dot_info__pb2._ZCELLULARMODULEINFO
_CELLULARSTATUS.fields_by_name['sim_cards'].message_type = info_dot_info__pb2._ZSIMCARDINFO
_CELLULARSTATUS.fields_by_name['providers'].message_type = info_dot_info__pb2._ZCELLULARPROVIDER
DESCRIPTOR.message_types_by_name['LocalProfile'] = _LOCALPROFILE
DESCRIPTOR.message_types_by_name['RadioStatus'] = _RADIOSTATUS
DESCRIPTOR.message_types_by_name['CellularStatus'] = _CELLULARSTATUS
DESCRIPTOR.message_types_by_name['RadioConfig'] = _RADIOCONFIG
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

LocalProfile = _reflection.GeneratedProtocolMessageType('LocalProfile', (_message.Message,), {
  'DESCRIPTOR' : _LOCALPROFILE,
  '__module__' : 'profile.local_profile_pb2'
  # @@protoc_insertion_point(class_scope:org.lfedge.eve.profile.LocalProfile)
  })
_sym_db.RegisterMessage(LocalProfile)

RadioStatus = _reflection.GeneratedProtocolMessageType('RadioStatus', (_message.Message,), {
  'DESCRIPTOR' : _RADIOSTATUS,
  '__module__' : 'profile.local_profile_pb2'
  # @@protoc_insertion_point(class_scope:org.lfedge.eve.profile.RadioStatus)
  })
_sym_db.RegisterMessage(RadioStatus)

CellularStatus = _reflection.GeneratedProtocolMessageType('CellularStatus', (_message.Message,), {
  'DESCRIPTOR' : _CELLULARSTATUS,
  '__module__' : 'profile.local_profile_pb2'
  # @@protoc_insertion_point(class_scope:org.lfedge.eve.profile.CellularStatus)
  })
_sym_db.RegisterMessage(CellularStatus)

RadioConfig = _reflection.GeneratedProtocolMessageType('RadioConfig', (_message.Message,), {
  'DESCRIPTOR' : _RADIOCONFIG,
  '__module__' : 'profile.local_profile_pb2'
  # @@protoc_insertion_point(class_scope:org.lfedge.eve.profile.RadioConfig)
  })
_sym_db.RegisterMessage(RadioConfig)


DESCRIPTOR._options = None
# @@protoc_insertion_point(module_scope)