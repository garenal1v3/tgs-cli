package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	gosession "github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"

	"github.com/searchtgcli/tgs/internal/config"
	"github.com/searchtgcli/tgs/internal/storage"
)

var (
	defaultAPIID   = "2040"
	defaultAPIHash = "b18441a1ff607e10a989891a5462e627"
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

// New creates a new Client for the given profile, creating the profile directory if needed.
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

// Open opens an existing profile for read-only access. Returns nil client if the profile does not exist.
func Open(profileName string) (*Client, error) {
	dbPath := config.ProfileDir(profileName) + "/session.db"
	sess, err := storage.OpenSession(dbPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
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
	err := c.api.Run(ctx, func(ctx context.Context) error {
		return fn(ctx, c.api.API())
	})
	if err != nil {
		// gotd wraps callback errors with "callback: " prefix — strip it.
		const prefix = "callback: "
		msg := err.Error()
		if strings.HasPrefix(msg, prefix) {
			return fmt.Errorf("%s", msg[len(prefix):])
		}
	}
	return err
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

// CachePath returns the path to the peer cache database for the given profile.
func CachePath(profileName string) string {
	return filepath.Join(config.ProfileDir(profileName), "cache.db")
}

// StatsCachePath returns the path to the sources stats cache database for
// the given profile. A separate file from CachePath because BoltDB only
// permits one process to open a database file at a time.
func StatsCachePath(profileName string) string {
	return filepath.Join(config.ProfileDir(profileName), "stats_cache.db")
}

// Close releases the session database.
func (c *Client) Close() error {
	return c.session.Close()
}
