// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.13.0
// source: wallet.proto

package wallet

import (
	proto "github.com/golang/protobuf/proto"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type SendToManyRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//
	//A slice of the outputs that should be created in the transaction produced.
	Outputs []*TxOut `protobuf:"bytes,1,rep,name=outputs,proto3" json:"outputs,omitempty"`
	//
	//The number of millisatoshis per byte that should be used when crafting
	//this transaction.
	MillisatPerByte int64 `protobuf:"varint,2,opt,name=millisat_per_byte,json=millisatPerByte,proto3" json:"millisat_per_byte,omitempty"`
	// Optional: if true the transaction will be pushed to the network
	Push bool `protobuf:"varint,3,opt,name=push,proto3" json:"push,omitempty"`
}

func (x *SendToManyRequest) Reset() {
	*x = SendToManyRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_wallet_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendToManyRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendToManyRequest) ProtoMessage() {}

func (x *SendToManyRequest) ProtoReflect() protoreflect.Message {
	mi := &file_wallet_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendToManyRequest.ProtoReflect.Descriptor instead.
func (*SendToManyRequest) Descriptor() ([]byte, []int) {
	return file_wallet_proto_rawDescGZIP(), []int{0}
}

func (x *SendToManyRequest) GetOutputs() []*TxOut {
	if x != nil {
		return x.Outputs
	}
	return nil
}

func (x *SendToManyRequest) GetMillisatPerByte() int64 {
	if x != nil {
		return x.MillisatPerByte
	}
	return 0
}

func (x *SendToManyRequest) GetPush() bool {
	if x != nil {
		return x.Push
	}
	return false
}

type SendToManyReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//
	//The serialized transaction sent out on the network.
	RawTx []byte `protobuf:"bytes,1,opt,name=raw_tx,json=rawTx,proto3" json:"raw_tx,omitempty"`
}

func (x *SendToManyReply) Reset() {
	*x = SendToManyReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_wallet_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendToManyReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendToManyReply) ProtoMessage() {}

func (x *SendToManyReply) ProtoReflect() protoreflect.Message {
	mi := &file_wallet_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendToManyReply.ProtoReflect.Descriptor instead.
func (*SendToManyReply) Descriptor() ([]byte, []int) {
	return file_wallet_proto_rawDescGZIP(), []int{1}
}

func (x *SendToManyReply) GetRawTx() []byte {
	if x != nil {
		return x.RawTx
	}
	return nil
}

type WalletAddressRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *WalletAddressRequest) Reset() {
	*x = WalletAddressRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_wallet_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WalletAddressRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WalletAddressRequest) ProtoMessage() {}

func (x *WalletAddressRequest) ProtoReflect() protoreflect.Message {
	mi := &file_wallet_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WalletAddressRequest.ProtoReflect.Descriptor instead.
func (*WalletAddressRequest) Descriptor() ([]byte, []int) {
	return file_wallet_proto_rawDescGZIP(), []int{2}
}

type WalletAddressReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The confidential address encoded using a blech32 format.
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	// The blinding private key for the given address encoded in hex format
	Blinding string `protobuf:"bytes,2,opt,name=blinding,proto3" json:"blinding,omitempty"`
}

func (x *WalletAddressReply) Reset() {
	*x = WalletAddressReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_wallet_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WalletAddressReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WalletAddressReply) ProtoMessage() {}

func (x *WalletAddressReply) ProtoReflect() protoreflect.Message {
	mi := &file_wallet_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WalletAddressReply.ProtoReflect.Descriptor instead.
func (*WalletAddressReply) Descriptor() ([]byte, []int) {
	return file_wallet_proto_rawDescGZIP(), []int{3}
}

func (x *WalletAddressReply) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *WalletAddressReply) GetBlinding() string {
	if x != nil {
		return x.Blinding
	}
	return ""
}

type BalanceInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The balance of the wallet
	TotalBalance uint64 `protobuf:"varint,1,opt,name=total_balance,json=totalBalance,proto3" json:"total_balance,omitempty"`
	// The confirmed balance of a wallet(with >= 1 confirmations)
	ConfirmedBalance uint64 `protobuf:"varint,2,opt,name=confirmed_balance,json=confirmedBalance,proto3" json:"confirmed_balance,omitempty"`
	// The unconfirmed balance of a wallet(with 0 confirmations)
	UnconfirmedBalance uint64 `protobuf:"varint,3,opt,name=unconfirmed_balance,json=unconfirmedBalance,proto3" json:"unconfirmed_balance,omitempty"`
}

func (x *BalanceInfo) Reset() {
	*x = BalanceInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_wallet_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BalanceInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BalanceInfo) ProtoMessage() {}

func (x *BalanceInfo) ProtoReflect() protoreflect.Message {
	mi := &file_wallet_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BalanceInfo.ProtoReflect.Descriptor instead.
func (*BalanceInfo) Descriptor() ([]byte, []int) {
	return file_wallet_proto_rawDescGZIP(), []int{4}
}

func (x *BalanceInfo) GetTotalBalance() uint64 {
	if x != nil {
		return x.TotalBalance
	}
	return 0
}

func (x *BalanceInfo) GetConfirmedBalance() uint64 {
	if x != nil {
		return x.ConfirmedBalance
	}
	return 0
}

func (x *BalanceInfo) GetUnconfirmedBalance() uint64 {
	if x != nil {
		return x.UnconfirmedBalance
	}
	return 0
}

type WalletBalanceRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *WalletBalanceRequest) Reset() {
	*x = WalletBalanceRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_wallet_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WalletBalanceRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WalletBalanceRequest) ProtoMessage() {}

func (x *WalletBalanceRequest) ProtoReflect() protoreflect.Message {
	mi := &file_wallet_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WalletBalanceRequest.ProtoReflect.Descriptor instead.
func (*WalletBalanceRequest) Descriptor() ([]byte, []int) {
	return file_wallet_proto_rawDescGZIP(), []int{5}
}

type WalletBalanceReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The balance info (total, confirmed, unconfirmed) of the wallet grouped by
	// asset
	Balance map[string]*BalanceInfo `protobuf:"bytes,1,rep,name=balance,proto3" json:"balance,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *WalletBalanceReply) Reset() {
	*x = WalletBalanceReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_wallet_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WalletBalanceReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WalletBalanceReply) ProtoMessage() {}

func (x *WalletBalanceReply) ProtoReflect() protoreflect.Message {
	mi := &file_wallet_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WalletBalanceReply.ProtoReflect.Descriptor instead.
func (*WalletBalanceReply) Descriptor() ([]byte, []int) {
	return file_wallet_proto_rawDescGZIP(), []int{6}
}

func (x *WalletBalanceReply) GetBalance() map[string]*BalanceInfo {
	if x != nil {
		return x.Balance
	}
	return nil
}

type TxOut struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The asset being spent
	Asset string `protobuf:"bytes,1,opt,name=asset,proto3" json:"asset,omitempty"`
	// The value of the output being spent.
	Value int64 `protobuf:"varint,2,opt,name=value,proto3" json:"value,omitempty"`
	// The confidential address of the output being spent.
	Address string `protobuf:"bytes,3,opt,name=address,proto3" json:"address,omitempty"`
}

func (x *TxOut) Reset() {
	*x = TxOut{}
	if protoimpl.UnsafeEnabled {
		mi := &file_wallet_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TxOut) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TxOut) ProtoMessage() {}

func (x *TxOut) ProtoReflect() protoreflect.Message {
	mi := &file_wallet_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TxOut.ProtoReflect.Descriptor instead.
func (*TxOut) Descriptor() ([]byte, []int) {
	return file_wallet_proto_rawDescGZIP(), []int{7}
}

func (x *TxOut) GetAsset() string {
	if x != nil {
		return x.Asset
	}
	return ""
}

func (x *TxOut) GetValue() int64 {
	if x != nil {
		return x.Value
	}
	return 0
}

func (x *TxOut) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

var File_wallet_proto protoreflect.FileDescriptor

var file_wallet_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x77, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x75,
	0x0a, 0x11, 0x53, 0x65, 0x6e, 0x64, 0x54, 0x6f, 0x4d, 0x61, 0x6e, 0x79, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x20, 0x0a, 0x07, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x06, 0x2e, 0x54, 0x78, 0x4f, 0x75, 0x74, 0x52, 0x07, 0x6f, 0x75,
	0x74, 0x70, 0x75, 0x74, 0x73, 0x12, 0x2a, 0x0a, 0x11, 0x6d, 0x69, 0x6c, 0x6c, 0x69, 0x73, 0x61,
	0x74, 0x5f, 0x70, 0x65, 0x72, 0x5f, 0x62, 0x79, 0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x0f, 0x6d, 0x69, 0x6c, 0x6c, 0x69, 0x73, 0x61, 0x74, 0x50, 0x65, 0x72, 0x42, 0x79, 0x74,
	0x65, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x75, 0x73, 0x68, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x04, 0x70, 0x75, 0x73, 0x68, 0x22, 0x28, 0x0a, 0x0f, 0x53, 0x65, 0x6e, 0x64, 0x54, 0x6f, 0x4d,
	0x61, 0x6e, 0x79, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x15, 0x0a, 0x06, 0x72, 0x61, 0x77, 0x5f,
	0x74, 0x78, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x72, 0x61, 0x77, 0x54, 0x78, 0x22,
	0x16, 0x0a, 0x14, 0x57, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x4a, 0x0a, 0x12, 0x57, 0x61, 0x6c, 0x6c, 0x65,
	0x74, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x18, 0x0a,
	0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x62, 0x6c, 0x69, 0x6e, 0x64,
	0x69, 0x6e, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x62, 0x6c, 0x69, 0x6e, 0x64,
	0x69, 0x6e, 0x67, 0x22, 0x90, 0x01, 0x0a, 0x0b, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x49,
	0x6e, 0x66, 0x6f, 0x12, 0x23, 0x0a, 0x0d, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x5f, 0x62, 0x61, 0x6c,
	0x61, 0x6e, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0c, 0x74, 0x6f, 0x74, 0x61,
	0x6c, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x12, 0x2b, 0x0a, 0x11, 0x63, 0x6f, 0x6e, 0x66,
	0x69, 0x72, 0x6d, 0x65, 0x64, 0x5f, 0x62, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x04, 0x52, 0x10, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x72, 0x6d, 0x65, 0x64, 0x42, 0x61,
	0x6c, 0x61, 0x6e, 0x63, 0x65, 0x12, 0x2f, 0x0a, 0x13, 0x75, 0x6e, 0x63, 0x6f, 0x6e, 0x66, 0x69,
	0x72, 0x6d, 0x65, 0x64, 0x5f, 0x62, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x12, 0x75, 0x6e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x72, 0x6d, 0x65, 0x64, 0x42,
	0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x22, 0x16, 0x0a, 0x14, 0x57, 0x61, 0x6c, 0x6c, 0x65, 0x74,
	0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x9a,
	0x01, 0x0a, 0x12, 0x57, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65,
	0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x3a, 0x0a, 0x07, 0x62, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x57, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x42,
	0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x2e, 0x42, 0x61, 0x6c, 0x61,
	0x6e, 0x63, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x62, 0x61, 0x6c, 0x61, 0x6e, 0x63,
	0x65, 0x1a, 0x48, 0x0a, 0x0c, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x6b, 0x65, 0x79, 0x12, 0x22, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x49, 0x6e, 0x66, 0x6f,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x4d, 0x0a, 0x05, 0x54,
	0x78, 0x4f, 0x75, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x61, 0x73, 0x73, 0x65, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x61, 0x73, 0x73, 0x65, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x32, 0xb6, 0x01, 0x0a, 0x06, 0x57,
	0x61, 0x6c, 0x6c, 0x65, 0x74, 0x12, 0x3b, 0x0a, 0x0d, 0x57, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x41,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x15, 0x2e, 0x57, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x41,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e,
	0x57, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x52, 0x65, 0x70,
	0x6c, 0x79, 0x12, 0x3b, 0x0a, 0x0d, 0x57, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x42, 0x61, 0x6c, 0x61,
	0x6e, 0x63, 0x65, 0x12, 0x15, 0x2e, 0x57, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x42, 0x61, 0x6c, 0x61,
	0x6e, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e, 0x57, 0x61, 0x6c,
	0x6c, 0x65, 0x74, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12,
	0x32, 0x0a, 0x0a, 0x53, 0x65, 0x6e, 0x64, 0x54, 0x6f, 0x4d, 0x61, 0x6e, 0x79, 0x12, 0x12, 0x2e,
	0x53, 0x65, 0x6e, 0x64, 0x54, 0x6f, 0x4d, 0x61, 0x6e, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x10, 0x2e, 0x53, 0x65, 0x6e, 0x64, 0x54, 0x6f, 0x4d, 0x61, 0x6e, 0x79, 0x52, 0x65,
	0x70, 0x6c, 0x79, 0x42, 0x42, 0x5a, 0x40, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x74, 0x64, 0x65, 0x78, 0x2d, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2f, 0x74,
	0x64, 0x65, 0x78, 0x2d, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2f, 0x61, 0x70, 0x69, 0x2d, 0x73,
	0x70, 0x65, 0x63, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x67, 0x65, 0x6e,
	0x2f, 0x77, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_wallet_proto_rawDescOnce sync.Once
	file_wallet_proto_rawDescData = file_wallet_proto_rawDesc
)

func file_wallet_proto_rawDescGZIP() []byte {
	file_wallet_proto_rawDescOnce.Do(func() {
		file_wallet_proto_rawDescData = protoimpl.X.CompressGZIP(file_wallet_proto_rawDescData)
	})
	return file_wallet_proto_rawDescData
}

var file_wallet_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_wallet_proto_goTypes = []interface{}{
	(*SendToManyRequest)(nil),    // 0: SendToManyRequest
	(*SendToManyReply)(nil),      // 1: SendToManyReply
	(*WalletAddressRequest)(nil), // 2: WalletAddressRequest
	(*WalletAddressReply)(nil),   // 3: WalletAddressReply
	(*BalanceInfo)(nil),          // 4: BalanceInfo
	(*WalletBalanceRequest)(nil), // 5: WalletBalanceRequest
	(*WalletBalanceReply)(nil),   // 6: WalletBalanceReply
	(*TxOut)(nil),                // 7: TxOut
	nil,                          // 8: WalletBalanceReply.BalanceEntry
}
var file_wallet_proto_depIdxs = []int32{
	7, // 0: SendToManyRequest.outputs:type_name -> TxOut
	8, // 1: WalletBalanceReply.balance:type_name -> WalletBalanceReply.BalanceEntry
	4, // 2: WalletBalanceReply.BalanceEntry.value:type_name -> BalanceInfo
	2, // 3: Wallet.WalletAddress:input_type -> WalletAddressRequest
	5, // 4: Wallet.WalletBalance:input_type -> WalletBalanceRequest
	0, // 5: Wallet.SendToMany:input_type -> SendToManyRequest
	3, // 6: Wallet.WalletAddress:output_type -> WalletAddressReply
	6, // 7: Wallet.WalletBalance:output_type -> WalletBalanceReply
	1, // 8: Wallet.SendToMany:output_type -> SendToManyReply
	6, // [6:9] is the sub-list for method output_type
	3, // [3:6] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_wallet_proto_init() }
func file_wallet_proto_init() {
	if File_wallet_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_wallet_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendToManyRequest); i {
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
		file_wallet_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendToManyReply); i {
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
		file_wallet_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WalletAddressRequest); i {
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
		file_wallet_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WalletAddressReply); i {
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
		file_wallet_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BalanceInfo); i {
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
		file_wallet_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WalletBalanceRequest); i {
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
		file_wallet_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WalletBalanceReply); i {
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
		file_wallet_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TxOut); i {
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
			RawDescriptor: file_wallet_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_wallet_proto_goTypes,
		DependencyIndexes: file_wallet_proto_depIdxs,
		MessageInfos:      file_wallet_proto_msgTypes,
	}.Build()
	File_wallet_proto = out.File
	file_wallet_proto_rawDesc = nil
	file_wallet_proto_goTypes = nil
	file_wallet_proto_depIdxs = nil
}
