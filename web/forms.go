package web

import "encoding/gob"

func init() {
	gob.Register(CreatePostForm{})
	gob.Register(CreateThreadForm{})
	gob.Register(CreateCommentForm{})
	gob.Register(RegisterUserForm{})
	gob.Register(LoginUserForm{})
	gob.Register(FormErrors{})
}

type FormErrors map[string]string

type CreatePostForm struct {
	Title   string
	Content string

	Errors FormErrors
}

func (f *CreatePostForm) Validate() bool {
	f.Errors = FormErrors{}
	if f.Title == "" {
		f.Errors["Title"] = "Please enter a title."
	}
	if f.Content == "" {
		f.Errors["Content"] = "Please enter some text."
	}

	return len(f.Errors) == 0
}

type CreateThreadForm struct {
	Title       string
	Description string

	Errors FormErrors
}

func (f *CreateThreadForm) Validate() bool {
	f.Errors = FormErrors{}
	if f.Title == "" {
		f.Errors["Title"] = "Please enter a title."
	}
	if f.Description == "" {
		f.Errors["Description"] = "Please enter a description."
	}

	return len(f.Errors) == 0
}

type CreateCommentForm struct {
	Content string

	Errors FormErrors
}

func (f *CreateCommentForm) Validate() bool {
	f.Errors = FormErrors{}
	if f.Content == "" {
		f.Errors["Content"] = "Please enter a comment."
	}

	return len(f.Errors) == 0
}

type RegisterUserForm struct {
	Username        string
	Password        string
	PasswordConfirm string
	UsernameTaken   bool

	Errors FormErrors
}

func (f *RegisterUserForm) Validate() bool {
	f.Errors = FormErrors{}
	if f.Username == "" {
		f.Errors["Username"] = "Please enter a username."
	} else if f.UsernameTaken {
		f.Errors["Username"] = "This username is already taken."
	}
	if f.Password == "" {
		f.Errors["Password"] = "Please enter a password."
	} else if len(f.Password) < 8 {
		f.Errors["Password"] = "Password must be at least 8 characters long."
	}
	if f.PasswordConfirm == "" {
		f.Errors["PasswordConfirm"] = "Please enter the same password."
	} else if f.Password != f.PasswordConfirm {
		f.Errors["PasswordConfirm"] = "Password confirmation does not match."
	}

	return len(f.Errors) == 0
}

type LoginUserForm struct {
	Username                  string
	Password                  string
	InvalidUsernameOrPassword bool

	Errors FormErrors
}

func (f *LoginUserForm) Validate() bool {
	f.Errors = FormErrors{}
	if f.Username == "" {
		f.Errors["Username"] = "Please enter a username."
	} else if f.InvalidUsernameOrPassword {
		f.Errors["Username"] = "Username or password is incorrect."
	}
	if f.Password == "" {
		f.Errors["Password"] = "Please enter a password."
	}

	return len(f.Errors) == 0
}
