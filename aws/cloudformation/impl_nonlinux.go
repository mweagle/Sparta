// +build !linux

package cloudformation

import (
	"os/user"
)

func platformUserName() string {
	currentUser, currentUserErr := user.Current()
	if nil != currentUserErr {
		return defaultUserName()
	}
	return currentUser.Username
}
