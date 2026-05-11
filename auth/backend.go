package auth

import (
	"context"
	"fmt"

	"github.com/iMerica/jango/orm"
)

type ModelBackend struct{}

func (mb *ModelBackend) Authenticate(ctx context.Context, username, password string) (*User, error) {
	db := orm.DefaultDB()
	if db == nil {
		return nil, fmt.Errorf("auth: database not available")
	}

	userModel := GetUserModel()
	pkCol := userModel.PKColumn()
	usernameCol := userModel.DBColumnForField("Username")
	passwordCol := userModel.DBColumnForField("Password")

	query := fmt.Sprintf("SELECT %s, %s, %s FROM %s WHERE %s = $1",
		pkCol, usernameCol, passwordCol,
		userModel.TableName, usernameCol)

	row := db.QueryRow(ctx, query, username)

	var id int64
	var uname, hashedPw string
	if err := row.Scan(&id, &uname, &hashedPw); err != nil {
		return nil, ErrAuthenticationFailed
	}

	if !CheckPassword(password, hashedPw) {
		return nil, ErrAuthenticationFailed
	}

	user := &User{
		ID:       id,
		Username: uname,
		Password: hashedPw,
	}

	if err := mb.loadUserFields(ctx, user, userModel, pkCol); err != nil {
		return user, nil
	}

	return user, nil
}

func (mb *ModelBackend) GetUser(ctx context.Context, userID int64) (*User, error) {
	db := orm.DefaultDB()
	if db == nil {
		return nil, fmt.Errorf("auth: database not available")
	}

	userModel := GetUserModel()

	query := fmt.Sprintf("SELECT id, username, email, password, first_name, last_name, is_active, is_staff, is_admin, last_login, date_joined FROM %s WHERE %s = $1",
		userModel.TableName, userModel.PKColumn())

	row := db.QueryRow(ctx, query, userID)
	user := &User{}
	var firstName, lastName *string
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &firstName, &lastName,
		&user.IsActive, &user.IsStaff, &user.IsAdmin, &user.LastLogin, &user.DateJoined); err != nil {
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

func (mb *ModelBackend) loadUserFields(ctx context.Context, user *User, userModel *orm.ModelMeta, pkCol string) error {
	db := orm.DefaultDB()
	if db == nil {
		return nil
	}

	query := fmt.Sprintf("SELECT email, first_name, last_name, is_active, is_staff, is_admin, last_login, date_joined FROM %s WHERE %s = $1",
		userModel.TableName, pkCol)

	row := db.QueryRow(ctx, query, user.ID)
	var firstName, lastName *string
	if err := row.Scan(&user.Email, &firstName, &lastName, &user.IsActive, &user.IsStaff, &user.IsAdmin,
		&user.LastLogin, &user.DateJoined); err != nil {
		return err
	}
	if firstName != nil {
		user.FirstName = *firstName
	}
	if lastName != nil {
		user.LastName = *lastName
	}

	return nil
}

func toBool(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case int64:
		return val != 0
	case string:
		return val == "true" || val == "1" || val == "t"
	default:
		return false
	}
}
