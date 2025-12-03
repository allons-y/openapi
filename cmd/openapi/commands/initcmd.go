// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package commands

import "github.com/allons-y/openapi/cmd/openapi/commands/initcmd"

// InitCmd is a command namespace for initializing things like a swagger spec.
type InitCmd struct {
	Model *initcmd.Spec `command:"spec"`
}

// Execute provides default empty implementation.
func (i *InitCmd) Execute(_ []string) error {
	return nil
}
