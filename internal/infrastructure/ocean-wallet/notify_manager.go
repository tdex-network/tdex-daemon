package oceanwallet

import (
	"context"
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"
	pb "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/ocean/v1"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"google.golang.org/grpc"
)

type notifyManager struct {
	client              pb.NotificationServiceClient
	chTxNotifications   chan ports.WalletTxNotification
	chUtxoNotifications chan ports.WalletUtxoNotification
}

func newNotifyManager(conn *grpc.ClientConn) (*notifyManager, error) {
	svc := &notifyManager{
		client:              pb.NewNotificationServiceClient(conn),
		chTxNotifications:   make(chan ports.WalletTxNotification),
		chUtxoNotifications: make(chan ports.WalletUtxoNotification),
	}

	if err := svc.startListeningForTxNotifications(); err != nil {
		return nil, fmt.Errorf(
			"failed to open stream for tx notifications: %s", err,
		)
	}
	if err := svc.startListeningForUtxoNotifications(); err != nil {
		return nil, fmt.Errorf(
			"failed to open stream for utxo notifications: %s", err,
		)
	}

	return svc, nil
}

func (m *notifyManager) GetTxNotifications() chan ports.WalletTxNotification {
	return m.chTxNotifications
}

func (m *notifyManager) GetUtxoNotifications() chan ports.WalletUtxoNotification {
	return m.chUtxoNotifications
}

func (m *notifyManager) startListeningForTxNotifications() error {
	stream, err := m.client.TransactionNotifications(
		context.Background(), &pb.TransactionNotificationsRequest{},
	)
	if err != nil {
		return err
	}

	go func() {
		for {
			notification, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					return
				}
				log.WithError(err).Warn("closed connection with ocean wallet's tx notification stream")
				return
			}

			select {
			case m.chTxNotifications <- txNotifyInfo{notification}:
				continue
			default:
			}
		}
	}()

	return nil
}

func (m *notifyManager) startListeningForUtxoNotifications() error {
	stream, err := m.client.UtxosNotifications(
		context.Background(), &pb.UtxosNotificationsRequest{},
	)
	if err != nil {
		return err
	}

	go func() {
		for {
			notification, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					return
				}
				log.WithError(err).Warn(
					"closed connection with ocean wallet's utxo notification stream",
				)
				return
			}

			select {
			case m.chUtxoNotifications <- utxoNotifyInfo{notification}:
				continue
			default:
			}
		}
	}()

	return nil
}

type txNotifyInfo struct {
	*pb.TransactionNotificationsResponse
}

func (i txNotifyInfo) GetEventType() ports.WalletTxEventType {
	return txEventType(i.TransactionNotificationsResponse.GetEventType())
}
func (i txNotifyInfo) GetTxHex() string {
	return i.TransactionNotificationsResponse.GetTxhex()
}
func (i txNotifyInfo) GetBlockDetails() ports.BlockInfo {
	return i.TransactionNotificationsResponse.GetBlockDetails()
}

type txEventType pb.TxEventType

func (t txEventType) IsUnconfirmed() bool {
	return int(t) == int(pb.TxEventType_TX_EVENT_TYPE_UNCONFIRMED)
}
func (t txEventType) IsConfirmed() bool {
	return int(t) == int(pb.TxEventType_TX_EVENT_TYPE_CONFIRMED)
}
func (t txEventType) IsBroadcasted() bool {
	return int(t) == int(pb.TxEventType_TX_EVENT_TYPE_BROADCASTED)
}

type utxoNotifyInfo struct {
	*pb.UtxosNotificationsResponse
}

func (i utxoNotifyInfo) GetEventType() ports.WalletUtxoEventType {
	return utxoEventType(i.UtxosNotificationsResponse.GetEventType())
}
func (i utxoNotifyInfo) GetUtxos() []ports.Utxo {
	utxos := make([]ports.Utxo, 0, len(i.UtxosNotificationsResponse.GetUtxos()))
	for _, u := range i.UtxosNotificationsResponse.GetUtxos() {
		utxos = append(utxos, utxoInfo{u})
	}
	return utxos
}
func (i utxoNotifyInfo) GetBlockDetails() ports.BlockInfo {
	// utxo := i.UtxosNotificationsResponse.GetUtxos()[0]
	// if status := utxo.GetSpentStatus(); status != nil {
	// 	return status.GetBlockInfo()
	// }

	// return utxo.GetConfirmedStatus().GetBlockInfo()
	return nil
}

type utxoEventType pb.UtxoEventType

func (t utxoEventType) IsUnconfirmed() bool {
	return int(t) == int(pb.UtxoEventType_UTXO_EVENT_TYPE_NEW)
}
func (t utxoEventType) IsSpent() bool {
	return int(t) == int(pb.UtxoEventType_UTXO_EVENT_TYPE_SPENT)
}
func (t utxoEventType) IsConfirmed() bool {
	return int(t) == int(pb.UtxoEventType_UTXO_EVENT_TYPE_CONFIRMED)
}
func (t utxoEventType) IsLocked() bool {
	return int(t) == int(pb.UtxoEventType_UTXO_EVENT_TYPE_LOCKED)
}
func (t utxoEventType) IsUnlocked() bool {
	return int(t) == int(pb.UtxoEventType_UTXO_EVENT_TYPE_UNLOCKED)
}

type utxoInfo struct {
	*pb.Utxo
}

func (i utxoInfo) GetConfirmedStatus() ports.UtxoStatus {
	return utxoStatusInfo{i.Utxo.GetConfirmedStatus()}
}

func (i utxoInfo) GetSpentStatus() ports.UtxoStatus {
	return utxoStatusInfo{i.Utxo.GetSpentStatus()}
}

type utxoStatusInfo struct {
	*pb.UtxoStatus
}

func (i utxoStatusInfo) GetBlockInfo() ports.BlockInfo {
	return i.UtxoStatus.GetBlockInfo()
}
