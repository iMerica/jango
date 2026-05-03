package auth

import (
	"context"
	"time"

	"github.com/iMerica/jango/orm"
	"github.com/iMerica/jango/signals"
)

func Login(ctx context.Context, user *User) error {
	now := time.Now()
	user.LastLogin = &now

	db := orm.DefaultDB()
	if db != nil {
		_, err := db.Exec(ctx,
			"UPDATE auth_user SET last_login = $1 WHERE id = $2",
			now, user.ID)
		if err != nil {
			return err
		}
	}

	signals.PostSave.Send(user, map[string]interface{}{
		"created": false,
		"model":   "auth.User",
	})

	return nil
}

func Logout(ctx context.Context, user *User) {
}

func CreateUser(ctx context.Context, username, email, password string) (*User, error) {
	hashedPassword, err := MakePassword(password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &User{
		Username:   username,
		Email:      email,
		Password:   hashedPassword,
		IsActive:   true,
		DateJoined: now,
	}

	db := orm.DefaultDB()
	if db == nil {
		return nil, orm.ErrDBNotAvailable
	}

	var id int64
	err = orm.Atomic(ctx, db, func(ctx context.Context) error {
		row := db.QueryRow(ctx,
			"INSERT INTO auth_user (username, email, password, is_active, is_staff, is_admin, date_joined) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
			username, email, hashedPassword, true, false, false, now)
		if err := row.Scan(&id); err != nil {
			return err
		}
		user.ID = id
		return nil
	})
	if err != nil {
		return nil, err
	}

	signals.PostSave.Send(user, map[string]interface{}{
		"created": true,
		"model":   "auth.User",
	})

	return user, nil
}

func CreateSuperUser(ctx context.Context, username, email, password string) (*User, error) {
	user, err := CreateUser(ctx, username, email, password)
	if err != nil {
		return nil, err
	}

	db := orm.DefaultDB()
	if db != nil {
		_, err = db.Exec(ctx,
			"UPDATE auth_user SET is_staff = $1, is_admin = $2 WHERE id = $3",
			true, true, user.ID)
		if err != nil {
			return nil, err
		}
	}

	user.IsStaff = true
	user.IsAdmin = true
	user.IsSuperuser = true
	return user, nil
}

func GetUserByID(ctx context.Context, userID int64) (*User, error) {
	db := orm.DefaultDB()
	if db == nil {
		return nil, orm.ErrDBNotAvailable
	}

	row := db.QueryRow(ctx,
		"SELECT id, username, email, password, first_name, last_name, is_active, is_staff, is_admin, date_joined FROM auth_user WHERE id = $1",
		userID)

	user := &User{}
	var lastLogin *string
	var dateJoined time.Time
	var firstName, lastName *string

	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password,
		&firstName, &lastName, &user.IsActive, &user.IsStaff, &user.IsAdmin, &dateJoined)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if firstName != nil {
		user.FirstName = *firstName
	}
	if lastName != nil {
		user.LastName = *lastName
	}
	user.DateJoined = dateJoined

	return user, nil
}

func GetUserByUsername(ctx context.Context, username string) (*User, error) {
	db := orm.DefaultDB()
	if db == nil {
		return nil, orm.ErrDBNotAvailable
	}

	row := db.QueryRow(ctx,
		"SELECT id, username, email, password, first_name, last_name, is_active, is_staff, is_admin FROM auth_user WHERE username = $1",
		username)

	user := &User{}
	var firstName, lastName *string

	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password,
		&firstName, &lastName, &user.IsActive, &user.IsStaff, &user.IsAdmin)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if firstName != nil {
		user.FirstName = *firstName
	}
	if lastName != nil {
		user.LastName = *lastName
	}

	return user, nil
}
