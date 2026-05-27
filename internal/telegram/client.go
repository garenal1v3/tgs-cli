package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	gosession "github.com/gotd/td/session"

	"github.com/searchtgcli/tgs/internal/config"
	"github.com/searchtgcli/tgs/internal/storage"
)

var (
	defaultAPIID   = "0"
	defaultAPIHash = ""
)

// UserInfo holds basic info about the authenticated user.
type UserInfo struct {
	ID        int64  `json:"id"`
	Phone     string `json:"phone"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// Client wraps gotd/td telegram.Client with session storage.
type Client struct {
	api     *telegram.Client
	session *storage.Session
}

// apiCredentials returns the API ID and hash, preferring env vars over defaults.
func apiCredentials() (int, string) {
	idStr := defaultAPIID
	hash := defaultAPIHash

	if envID := os.Getenv("TGS_API_ID"); envID != "" {
		idStr = envID
	}
	if envHash := os.Getenv("TGS_API_HASH"); envHash != "" {
		hash = envHash
	}

	id, _ := strconv.Atoi(idStr)
	return id, hash
}

// New creates a new Client for the given profile.
func New(profileName string) (*Client, error) {
	dbPath := config.ProfileDir(profileName) + "/session.db"
	sess, err := storage.NewSession(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open session storage: %w", err)
	}

	apiID, apiHash := apiCredentials()
	api := telegram.NewClient(apiID, apiHash, telegram.Options{
		SessionStorage: sess,
	})

	return &Client{api: api, session: sess}, nil
}

// Run connects to Telegram and executes fn within the authenticated session.
func (c *Client) Run(ctx context.Context, fn func(ctx context.Context, api *tg.Client) error) error {
	return c.api.Run(ctx, func(ctx context.Context) error {
		return fn(ctx, c.api.API())
	})
}

// RawClient returns the underlying gotd/td telegram.Client.
func (c *Client) RawClient() *telegram.Client {
	return c.api
}

// StoreMeta persists UserInfo and individual fields to the session store.
func (c *Client) StoreMeta(info *UserInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	if err := c.session.StoreMeta("user_info", data); err != nil {
		return err
	}
	if info.Phone != "" {
		c.session.StoreMeta("phone", []byte(info.Phone)) //nolint:errcheck
	}
	if info.Username != "" {
		c.session.StoreMeta("username", []byte(info.Username)) //nolint:errcheck
	}
	if info.ID != 0 {
		c.session.StoreMeta("user_id", []byte(strconv.FormatInt(info.ID, 10))) //nolint:errcheck
	}
	return nil
}

// LoadMeta retrieves UserInfo from the session store.
func (c *Client) LoadMeta() (*UserInfo, error) {
	data, err := c.session.LoadMeta("user_info")
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	var info UserInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// SessionStorage returns the session.Storage used by this client.
// This is needed for MethodDesktop authentication, which must write to the
// same storage that was passed to telegram.NewClient.
func (c *Client) SessionStorage() gosession.Storage {
	return c.session
}

// Close releases the session database.
func (c *Client) Close() error {
	return c.session.Close()
}
