package sources

import (
	"context"

	"github.com/gotd/td/tg"
)

// API is the narrow subset of *tg.Client used by Service. Defined as an
// interface so tests can supply a mock without spinning up a real connection.
type API interface {
	MessagesGetDialogs(ctx context.Context, req *tg.MessagesGetDialogsRequest) (tg.MessagesDialogsClass, error)
	MessagesSearch(ctx context.Context, req *tg.MessagesSearchRequest) (tg.MessagesMessagesClass, error)
	MessagesGetHistory(ctx context.Context, req *tg.MessagesGetHistoryRequest) (tg.MessagesMessagesClass, error)
	ChannelsGetFullChannel(ctx context.Context, ch tg.InputChannelClass) (*tg.MessagesChatFull, error)
	MessagesGetFullChat(ctx context.Context, chatID int64) (*tg.MessagesChatFull, error)
	UsersGetFullUser(ctx context.Context, u tg.InputUserClass) (*tg.UsersUserFull, error)
	MessagesGetDialogFilters(ctx context.Context) (*tg.MessagesDialogFilters, error)
}
