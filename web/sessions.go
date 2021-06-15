package web

import (
	"context"
	"database/sql"
	"encoding/gob"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/blrobin2/goreddit"
	"github.com/google/uuid"
)

func init() {
	gob.Register(uuid.UUID{})
}

func NewSessionManager(dataSourceName string) (*scs.SessionManager, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}

	sessions := scs.New()
	sessions.Store = postgresstore.New(db)

	return sessions, nil
}

type SessionData struct {
	FlashMessage string
	Form         interface{}
	User         goreddit.User
	CSRFToken    string
	LoggedIn     bool
}

func GetSessionData(session *scs.SessionManager, ctx context.Context) SessionData {
	var data SessionData

	data.FlashMessage = session.PopString(ctx, "flash")
	data.User, data.LoggedIn = ctx.Value(KeyUserID).(goreddit.User)

	data.Form = session.Pop(ctx, "form")
	if data.Form == nil {
		data.Form = map[string]string{}
	}

	return data
}
