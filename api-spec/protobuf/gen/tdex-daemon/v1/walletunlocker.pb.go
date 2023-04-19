// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        (unknown)
// source: tdex-daemon/v1/walletunlocker.proto

package tdex_daemonv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type InitWalletResponse_Status int32

const (
	InitWalletResponse_STATUS_PROCESSING InitWalletResponse_Status = 0
	InitWalletResponse_STATUS_DONE       InitWalletResponse_Status = 1
)

// Enum value maps for InitWalletResponse_Status.
var (
	InitWalletResponse_Status_name = map[int32]string{
		0: "STATUS_PROCESSING",
		1: "STATUS_DONE",
	}
	InitWalletResponse_Status_value = map[string]int32{
		"STATUS_PROCESSING": 0,
		"STATUS_DONE":       1,
	}
)

func (x InitWalletResponse_Status) Enum() *InitWalletResponse_Status {
	p := new(InitWalletResponse_Status)
	*p = x
	return p
}

func (x InitWalletResponse_Status) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (InitWalletResponse_Status) Descriptor() protoreflect.EnumDescriptor {
	return file_tdex_daemon_v1_walletunlocker_proto_enumTypes[0].Descriptor()
}

func (InitWalletResponse_Status) Type() protoreflect.EnumType {
	return &file_tdex_daemon_v1_walletunlocker_proto_enumTypes[0]
}

func (x InitWalletResponse_Status) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use InitWalletResponse_Status.Descriptor instead.
func (InitWalletResponse_Status) EnumDescriptor() ([]byte, []int) {
	return file_tdex_daemon_v1_walletunlocker_proto_rawDescGZIP(), []int{3, 0}
}

type GenSeedRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GenSeedRequest) Reset() {
	*x = GenSeedRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tdex_daemon_v1_walletunlocker_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GenSeedRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GenSeedRequest) ProtoMessage() {}

func (x *GenSeedRequest) ProtoReflect() protoreflect.Message {
	mi := &file_tdex_daemon_v1_walletunlocker_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GenSeedRequest.ProtoReflect.Descriptor instead.
func (*GenSeedRequest) Descriptor() ([]byte, []int) {
	return file_tdex_daemon_v1_walletunlocker_proto_rawDescGZIP(), []int{0}
}

type GenSeedResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SeedMnemonic []string `protobuf:"bytes,1,rep,name=seed_mnemonic,json=seedMnemonic,proto3" json:"seed_mnemonic,omitempty"`
}

func (x *GenSeedResponse) Reset() {
	*x = GenSeedResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tdex_daemon_v1_walletunlocker_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GenSeedResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GenSeedResponse) ProtoMessage() {}

func (x *GenSeedResponse) ProtoReflect() protoreflect.Message {
	mi := &file_tdex_daemon_v1_walletunlocker_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GenSeedResponse.ProtoReflect.Descriptor instead.
func (*GenSeedResponse) Descriptor() ([]byte, []int) {
	return file_tdex_daemon_v1_walletunlocker_proto_rawDescGZIP(), []int{1}
}

func (x *GenSeedResponse) GetSeedMnemonic() []string {
	if x != nil {
		return x.SeedMnemonic
	}
	return nil
}

type InitWalletRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// wallet_password is the passphrase that should be used to encrypt the
	// wallet. This MUST be at least 8 chars in length. After creation, this
	// password is required to unlock the daemon.
	WalletPassword []byte `protobuf:"bytes,1,opt,name=wallet_password,json=walletPassword,proto3" json:"wallet_password,omitempty"`
	// seed_mnemonic is a 24-word mnemonic that encodes a prior seed obtained by the
	// user. This MUST be a generated by the GenSeed method
	SeedMnemonic []string `protobuf:"bytes,2,rep,name=seed_mnemonic,json=seedMnemonic,proto3" json:"seed_mnemonic,omitempty"`
	// the flag to let the daemon restore existing funds for the wallet.
	Restore bool `protobuf:"varint,3,opt,name=restore,proto3" json:"restore,omitempty"`
}

func (x *InitWalletRequest) Reset() {
	*x = InitWalletRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tdex_daemon_v1_walletunlocker_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InitWalletRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InitWalletRequest) ProtoMessage() {}

func (x *InitWalletRequest) ProtoReflect() protoreflect.Message {
	mi := &file_tdex_daemon_v1_walletunlocker_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InitWalletRequest.ProtoReflect.Descriptor instead.
func (*InitWalletRequest) Descriptor() ([]byte, []int) {
	return file_tdex_daemon_v1_walletunlocker_proto_rawDescGZIP(), []int{2}
}

func (x *InitWalletRequest) GetWalletPassword() []byte {
	if x != nil {
		return x.WalletPassword
	}
	return nil
}

func (x *InitWalletRequest) GetSeedMnemonic() []string {
	if x != nil {
		return x.SeedMnemonic
	}
	return nil
}

func (x *InitWalletRequest) GetRestore() bool {
	if x != nil {
		return x.Restore
	}
	return false
}

type InitWalletResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Account int32                     `protobuf:"varint,1,opt,name=account,proto3" json:"account,omitempty"`
	Status  InitWalletResponse_Status `protobuf:"varint,2,opt,name=status,proto3,enum=tdex_daemon.v1.InitWalletResponse_Status" json:"status,omitempty"`
	Data    string                    `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *InitWalletResponse) Reset() {
	*x = InitWalletResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tdex_daemon_v1_walletunlocker_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InitWalletResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InitWalletResponse) ProtoMessage() {}

func (x *InitWalletResponse) ProtoReflect() protoreflect.Message {
	mi := &file_tdex_daemon_v1_walletunlocker_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InitWalletResponse.ProtoReflect.Descriptor instead.
func (*InitWalletResponse) Descriptor() ([]byte, []int) {
	return file_tdex_daemon_v1_walletunlocker_proto_rawDescGZIP(), []int{3}
}

func (x *InitWalletResponse) GetAccount() int32 {
	if x != nil {
		return x.Account
	}
	return 0
}

func (x *InitWalletResponse) GetStatus() InitWalletResponse_Status {
	if x != nil {
		return x.Status
	}
	return InitWalletResponse_STATUS_PROCESSING
}

func (x *InitWalletResponse) GetData() string {
	if x != nil {
		return x.Data
	}
	return ""
}

type UnlockWalletRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// wallet_password should be the current valid passphrase for the daemon. This
	// will be required to decrypt on-disk material that the daemon requires to
	// function properly.
	WalletPassword []byte `protobuf:"bytes,1,opt,name=wallet_password,json=walletPassword,proto3" json:"wallet_password,omitempty"`
}

func (x *UnlockWalletRequest) Reset() {
	*x = UnlockWalletRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tdex_daemon_v1_walletunlocker_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UnlockWalletRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UnlockWalletRequest) ProtoMessage() {}

func (x *UnlockWalletRequest) ProtoReflect() protoreflect.Message {
	mi := &file_tdex_daemon_v1_walletunlocker_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UnlockWalletRequest.ProtoReflect.Descriptor instead.
func (*UnlockWalletRequest) Descriptor() ([]byte, []int) {
	return file_tdex_daemon_v1_walletunlocker_proto_rawDescGZIP(), []int{4}
}

func (x *UnlockWalletRequest) GetWalletPassword() []byte {
	if x != nil {
		return x.WalletPassword
	}
	return nil
}

type UnlockWalletResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *UnlockWalletResponse) Reset() {
	*x = UnlockWalletResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tdex_daemon_v1_walletunlocker_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UnlockWalletResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UnlockWalletResponse) ProtoMessage() {}

func (x *UnlockWalletResponse) ProtoReflect() protoreflect.Message {
	mi := &file_tdex_daemon_v1_walletunlocker_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UnlockWalletResponse.ProtoReflect.Descriptor instead.
func (*UnlockWalletResponse) Descriptor() ([]byte, []int) {
	return file_tdex_daemon_v1_walletunlocker_proto_rawDescGZIP(), []int{5}
}

type ChangePasswordRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// current_password should be the current valid passphrase used to unlock the
	// daemon.
	CurrentPassword []byte `protobuf:"bytes,1,opt,name=current_password,json=currentPassword,proto3" json:"current_password,omitempty"`
	// new_password should be the new passphrase that will be needed to unlock the
	// daemon.
	NewPassword []byte `protobuf:"bytes,2,opt,name=new_password,json=newPassword,proto3" json:"new_password,omitempty"`
}

func (x *ChangePasswordRequest) Reset() {
	*x = ChangePasswordRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tdex_daemon_v1_walletunlocker_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ChangePasswordRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChangePasswordRequest) ProtoMessage() {}

func (x *ChangePasswordRequest) ProtoReflect() protoreflect.Message {
	mi := &file_tdex_daemon_v1_walletunlocker_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChangePasswordRequest.ProtoReflect.Descriptor instead.
func (*ChangePasswordRequest) Descriptor() ([]byte, []int) {
	return file_tdex_daemon_v1_walletunlocker_proto_rawDescGZIP(), []int{6}
}

func (x *ChangePasswordRequest) GetCurrentPassword() []byte {
	if x != nil {
		return x.CurrentPassword
	}
	return nil
}

func (x *ChangePasswordRequest) GetNewPassword() []byte {
	if x != nil {
		return x.NewPassword
	}
	return nil
}

type ChangePasswordResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ChangePasswordResponse) Reset() {
	*x = ChangePasswordResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tdex_daemon_v1_walletunlocker_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ChangePasswordResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChangePasswordResponse) ProtoMessage() {}

func (x *ChangePasswordResponse) ProtoReflect() protoreflect.Message {
	mi := &file_tdex_daemon_v1_walletunlocker_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChangePasswordResponse.ProtoReflect.Descriptor instead.
func (*ChangePasswordResponse) Descriptor() ([]byte, []int) {
	return file_tdex_daemon_v1_walletunlocker_proto_rawDescGZIP(), []int{7}
}

type IsReadyRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *IsReadyRequest) Reset() {
	*x = IsReadyRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tdex_daemon_v1_walletunlocker_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IsReadyRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IsReadyRequest) ProtoMessage() {}

func (x *IsReadyRequest) ProtoReflect() protoreflect.Message {
	mi := &file_tdex_daemon_v1_walletunlocker_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IsReadyRequest.ProtoReflect.Descriptor instead.
func (*IsReadyRequest) Descriptor() ([]byte, []int) {
	return file_tdex_daemon_v1_walletunlocker_proto_rawDescGZIP(), []int{8}
}

type IsReadyResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// whether the daemon is initialized with an HD wallet.
	Initialized bool `protobuf:"varint,1,opt,name=initialized,proto3" json:"initialized,omitempty"`
	// whether the daemon's wallet is unlocked.
	Unlocked bool `protobuf:"varint,2,opt,name=unlocked,proto3" json:"unlocked,omitempty"`
	// whether the daemon's wallet utxo set is up-to-date'.
	Synced bool `protobuf:"varint,3,opt,name=synced,proto3" json:"synced,omitempty"`
}

func (x *IsReadyResponse) Reset() {
	*x = IsReadyResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tdex_daemon_v1_walletunlocker_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IsReadyResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IsReadyResponse) ProtoMessage() {}

func (x *IsReadyResponse) ProtoReflect() protoreflect.Message {
	mi := &file_tdex_daemon_v1_walletunlocker_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IsReadyResponse.ProtoReflect.Descriptor instead.
func (*IsReadyResponse) Descriptor() ([]byte, []int) {
	return file_tdex_daemon_v1_walletunlocker_proto_rawDescGZIP(), []int{9}
}

func (x *IsReadyResponse) GetInitialized() bool {
	if x != nil {
		return x.Initialized
	}
	return false
}

func (x *IsReadyResponse) GetUnlocked() bool {
	if x != nil {
		return x.Unlocked
	}
	return false
}

func (x *IsReadyResponse) GetSynced() bool {
	if x != nil {
		return x.Synced
	}
	return false
}

var File_tdex_daemon_v1_walletunlocker_proto protoreflect.FileDescriptor

var file_tdex_daemon_v1_walletunlocker_proto_rawDesc = []byte{
	0x0a, 0x23, 0x74, 0x64, 0x65, 0x78, 0x2d, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2f, 0x76, 0x31,
	0x2f, 0x77, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x75, 0x6e, 0x6c, 0x6f, 0x63, 0x6b, 0x65, 0x72, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0e, 0x74, 0x64, 0x65, 0x78, 0x5f, 0x64, 0x61, 0x65, 0x6d,
	0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x22, 0x10, 0x0a, 0x0e, 0x47, 0x65, 0x6e, 0x53, 0x65, 0x65, 0x64,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x36, 0x0a, 0x0f, 0x47, 0x65, 0x6e, 0x53, 0x65,
	0x65, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x23, 0x0a, 0x0d, 0x73, 0x65,
	0x65, 0x64, 0x5f, 0x6d, 0x6e, 0x65, 0x6d, 0x6f, 0x6e, 0x69, 0x63, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x0c, 0x73, 0x65, 0x65, 0x64, 0x4d, 0x6e, 0x65, 0x6d, 0x6f, 0x6e, 0x69, 0x63, 0x22,
	0x7b, 0x0a, 0x11, 0x49, 0x6e, 0x69, 0x74, 0x57, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x27, 0x0a, 0x0f, 0x77, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x5f, 0x70,
	0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e, 0x77,
	0x61, 0x6c, 0x6c, 0x65, 0x74, 0x50, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x12, 0x23, 0x0a,
	0x0d, 0x73, 0x65, 0x65, 0x64, 0x5f, 0x6d, 0x6e, 0x65, 0x6d, 0x6f, 0x6e, 0x69, 0x63, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x0c, 0x73, 0x65, 0x65, 0x64, 0x4d, 0x6e, 0x65, 0x6d, 0x6f, 0x6e,
	0x69, 0x63, 0x12, 0x18, 0x0a, 0x07, 0x72, 0x65, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x07, 0x72, 0x65, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x22, 0xb7, 0x01, 0x0a,
	0x12, 0x49, 0x6e, 0x69, 0x74, 0x57, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x41, 0x0a,
	0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x29, 0x2e,
	0x74, 0x64, 0x65, 0x78, 0x5f, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x49,
	0x6e, 0x69, 0x74, 0x57, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x64, 0x61, 0x74, 0x61, 0x22, 0x30, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x15,
	0x0a, 0x11, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x50, 0x52, 0x4f, 0x43, 0x45, 0x53, 0x53,
	0x49, 0x4e, 0x47, 0x10, 0x00, 0x12, 0x0f, 0x0a, 0x0b, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f,
	0x44, 0x4f, 0x4e, 0x45, 0x10, 0x01, 0x22, 0x3e, 0x0a, 0x13, 0x55, 0x6e, 0x6c, 0x6f, 0x63, 0x6b,
	0x57, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x27, 0x0a,
	0x0f, 0x77, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x5f, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e, 0x77, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x50, 0x61,
	0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x22, 0x16, 0x0a, 0x14, 0x55, 0x6e, 0x6c, 0x6f, 0x63, 0x6b,
	0x57, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x65,
	0x0a, 0x15, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x50, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x29, 0x0a, 0x10, 0x63, 0x75, 0x72, 0x72, 0x65,
	0x6e, 0x74, 0x5f, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x0f, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x50, 0x61, 0x73, 0x73, 0x77, 0x6f,
	0x72, 0x64, 0x12, 0x21, 0x0a, 0x0c, 0x6e, 0x65, 0x77, 0x5f, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f,
	0x72, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x6e, 0x65, 0x77, 0x50, 0x61, 0x73,
	0x73, 0x77, 0x6f, 0x72, 0x64, 0x22, 0x18, 0x0a, 0x16, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x50,
	0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x10, 0x0a, 0x0e, 0x49, 0x73, 0x52, 0x65, 0x61, 0x64, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x22, 0x67, 0x0a, 0x0f, 0x49, 0x73, 0x52, 0x65, 0x61, 0x64, 0x79, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x6c, 0x69,
	0x7a, 0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x69, 0x6e, 0x69, 0x74, 0x69,
	0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x75, 0x6e, 0x6c, 0x6f, 0x63, 0x6b,
	0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x75, 0x6e, 0x6c, 0x6f, 0x63, 0x6b,
	0x65, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x79, 0x6e, 0x63, 0x65, 0x64, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x06, 0x73, 0x79, 0x6e, 0x63, 0x65, 0x64, 0x32, 0xc2, 0x03, 0x0a, 0x15, 0x57,
	0x61, 0x6c, 0x6c, 0x65, 0x74, 0x55, 0x6e, 0x6c, 0x6f, 0x63, 0x6b, 0x65, 0x72, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x12, 0x4a, 0x0a, 0x07, 0x47, 0x65, 0x6e, 0x53, 0x65, 0x65, 0x64, 0x12,
	0x1e, 0x2e, 0x74, 0x64, 0x65, 0x78, 0x5f, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31,
	0x2e, 0x47, 0x65, 0x6e, 0x53, 0x65, 0x65, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x1f, 0x2e, 0x74, 0x64, 0x65, 0x78, 0x5f, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31,
	0x2e, 0x47, 0x65, 0x6e, 0x53, 0x65, 0x65, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x55, 0x0a, 0x0a, 0x49, 0x6e, 0x69, 0x74, 0x57, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x12, 0x21,
	0x2e, 0x74, 0x64, 0x65, 0x78, 0x5f, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e,
	0x49, 0x6e, 0x69, 0x74, 0x57, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x22, 0x2e, 0x74, 0x64, 0x65, 0x78, 0x5f, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2e,
	0x76, 0x31, 0x2e, 0x49, 0x6e, 0x69, 0x74, 0x57, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x30, 0x01, 0x12, 0x59, 0x0a, 0x0c, 0x55, 0x6e, 0x6c, 0x6f, 0x63,
	0x6b, 0x57, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x12, 0x23, 0x2e, 0x74, 0x64, 0x65, 0x78, 0x5f, 0x64,
	0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x6e, 0x6c, 0x6f, 0x63, 0x6b, 0x57,
	0x61, 0x6c, 0x6c, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x24, 0x2e, 0x74,
	0x64, 0x65, 0x78, 0x5f, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x6e,
	0x6c, 0x6f, 0x63, 0x6b, 0x57, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x5f, 0x0a, 0x0e, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x50, 0x61, 0x73, 0x73,
	0x77, 0x6f, 0x72, 0x64, 0x12, 0x25, 0x2e, 0x74, 0x64, 0x65, 0x78, 0x5f, 0x64, 0x61, 0x65, 0x6d,
	0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x50, 0x61, 0x73, 0x73,
	0x77, 0x6f, 0x72, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x26, 0x2e, 0x74, 0x64,
	0x65, 0x78, 0x5f, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x68, 0x61,
	0x6e, 0x67, 0x65, 0x50, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x4a, 0x0a, 0x07, 0x49, 0x73, 0x52, 0x65, 0x61, 0x64, 0x79, 0x12, 0x1e,
	0x2e, 0x74, 0x64, 0x65, 0x78, 0x5f, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e,
	0x49, 0x73, 0x52, 0x65, 0x61, 0x64, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1f,
	0x2e, 0x74, 0x64, 0x65, 0x78, 0x5f, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e,
	0x49, 0x73, 0x52, 0x65, 0x61, 0x64, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42,
	0xd6, 0x01, 0x0a, 0x12, 0x63, 0x6f, 0x6d, 0x2e, 0x74, 0x64, 0x65, 0x78, 0x5f, 0x64, 0x61, 0x65,
	0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x42, 0x13, 0x57, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x75, 0x6e,
	0x6c, 0x6f, 0x63, 0x6b, 0x65, 0x72, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x56, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x74, 0x64, 0x65, 0x78, 0x2d, 0x6e,
	0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2f, 0x74, 0x64, 0x65, 0x78, 0x2d, 0x64, 0x61, 0x65, 0x6d,
	0x6f, 0x6e, 0x2f, 0x61, 0x70, 0x69, 0x2d, 0x73, 0x70, 0x65, 0x63, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x74, 0x64, 0x65, 0x78, 0x2d, 0x64, 0x61,
	0x65, 0x6d, 0x6f, 0x6e, 0x2f, 0x76, 0x31, 0x3b, 0x74, 0x64, 0x65, 0x78, 0x5f, 0x64, 0x61, 0x65,
	0x6d, 0x6f, 0x6e, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x54, 0x58, 0x58, 0xaa, 0x02, 0x0d, 0x54, 0x64,
	0x65, 0x78, 0x44, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x0d, 0x54, 0x64,
	0x65, 0x78, 0x44, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x19, 0x54, 0x64,
	0x65, 0x78, 0x44, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x0e, 0x54, 0x64, 0x65, 0x78, 0x44, 0x61,
	0x65, 0x6d, 0x6f, 0x6e, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_tdex_daemon_v1_walletunlocker_proto_rawDescOnce sync.Once
	file_tdex_daemon_v1_walletunlocker_proto_rawDescData = file_tdex_daemon_v1_walletunlocker_proto_rawDesc
)

func file_tdex_daemon_v1_walletunlocker_proto_rawDescGZIP() []byte {
	file_tdex_daemon_v1_walletunlocker_proto_rawDescOnce.Do(func() {
		file_tdex_daemon_v1_walletunlocker_proto_rawDescData = protoimpl.X.CompressGZIP(file_tdex_daemon_v1_walletunlocker_proto_rawDescData)
	})
	return file_tdex_daemon_v1_walletunlocker_proto_rawDescData
}

var file_tdex_daemon_v1_walletunlocker_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_tdex_daemon_v1_walletunlocker_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_tdex_daemon_v1_walletunlocker_proto_goTypes = []interface{}{
	(InitWalletResponse_Status)(0), // 0: tdex_daemon.v1.InitWalletResponse.Status
	(*GenSeedRequest)(nil),         // 1: tdex_daemon.v1.GenSeedRequest
	(*GenSeedResponse)(nil),        // 2: tdex_daemon.v1.GenSeedResponse
	(*InitWalletRequest)(nil),      // 3: tdex_daemon.v1.InitWalletRequest
	(*InitWalletResponse)(nil),     // 4: tdex_daemon.v1.InitWalletResponse
	(*UnlockWalletRequest)(nil),    // 5: tdex_daemon.v1.UnlockWalletRequest
	(*UnlockWalletResponse)(nil),   // 6: tdex_daemon.v1.UnlockWalletResponse
	(*ChangePasswordRequest)(nil),  // 7: tdex_daemon.v1.ChangePasswordRequest
	(*ChangePasswordResponse)(nil), // 8: tdex_daemon.v1.ChangePasswordResponse
	(*IsReadyRequest)(nil),         // 9: tdex_daemon.v1.IsReadyRequest
	(*IsReadyResponse)(nil),        // 10: tdex_daemon.v1.IsReadyResponse
}
var file_tdex_daemon_v1_walletunlocker_proto_depIdxs = []int32{
	0,  // 0: tdex_daemon.v1.InitWalletResponse.status:type_name -> tdex_daemon.v1.InitWalletResponse.Status
	1,  // 1: tdex_daemon.v1.WalletUnlockerService.GenSeed:input_type -> tdex_daemon.v1.GenSeedRequest
	3,  // 2: tdex_daemon.v1.WalletUnlockerService.InitWallet:input_type -> tdex_daemon.v1.InitWalletRequest
	5,  // 3: tdex_daemon.v1.WalletUnlockerService.UnlockWallet:input_type -> tdex_daemon.v1.UnlockWalletRequest
	7,  // 4: tdex_daemon.v1.WalletUnlockerService.ChangePassword:input_type -> tdex_daemon.v1.ChangePasswordRequest
	9,  // 5: tdex_daemon.v1.WalletUnlockerService.IsReady:input_type -> tdex_daemon.v1.IsReadyRequest
	2,  // 6: tdex_daemon.v1.WalletUnlockerService.GenSeed:output_type -> tdex_daemon.v1.GenSeedResponse
	4,  // 7: tdex_daemon.v1.WalletUnlockerService.InitWallet:output_type -> tdex_daemon.v1.InitWalletResponse
	6,  // 8: tdex_daemon.v1.WalletUnlockerService.UnlockWallet:output_type -> tdex_daemon.v1.UnlockWalletResponse
	8,  // 9: tdex_daemon.v1.WalletUnlockerService.ChangePassword:output_type -> tdex_daemon.v1.ChangePasswordResponse
	10, // 10: tdex_daemon.v1.WalletUnlockerService.IsReady:output_type -> tdex_daemon.v1.IsReadyResponse
	6,  // [6:11] is the sub-list for method output_type
	1,  // [1:6] is the sub-list for method input_type
	1,  // [1:1] is the sub-list for extension type_name
	1,  // [1:1] is the sub-list for extension extendee
	0,  // [0:1] is the sub-list for field type_name
}

func init() { file_tdex_daemon_v1_walletunlocker_proto_init() }
func file_tdex_daemon_v1_walletunlocker_proto_init() {
	if File_tdex_daemon_v1_walletunlocker_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_tdex_daemon_v1_walletunlocker_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GenSeedRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_tdex_daemon_v1_walletunlocker_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GenSeedResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_tdex_daemon_v1_walletunlocker_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InitWalletRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_tdex_daemon_v1_walletunlocker_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InitWalletResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_tdex_daemon_v1_walletunlocker_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UnlockWalletRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_tdex_daemon_v1_walletunlocker_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UnlockWalletResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_tdex_daemon_v1_walletunlocker_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ChangePasswordRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_tdex_daemon_v1_walletunlocker_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ChangePasswordResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_tdex_daemon_v1_walletunlocker_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IsReadyRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_tdex_daemon_v1_walletunlocker_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IsReadyResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_tdex_daemon_v1_walletunlocker_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_tdex_daemon_v1_walletunlocker_proto_goTypes,
		DependencyIndexes: file_tdex_daemon_v1_walletunlocker_proto_depIdxs,
		EnumInfos:         file_tdex_daemon_v1_walletunlocker_proto_enumTypes,
		MessageInfos:      file_tdex_daemon_v1_walletunlocker_proto_msgTypes,
	}.Build()
	File_tdex_daemon_v1_walletunlocker_proto = out.File
	file_tdex_daemon_v1_walletunlocker_proto_rawDesc = nil
	file_tdex_daemon_v1_walletunlocker_proto_goTypes = nil
	file_tdex_daemon_v1_walletunlocker_proto_depIdxs = nil
}
