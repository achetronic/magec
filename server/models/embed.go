// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed wakeword
var wakeword embed.FS

//go:embed auxiliary
var auxiliary embed.FS

// WakewordFS returns a filesystem rooted at the wakeword models directory.
func WakewordFS() (fs.FS, error) {
	return fs.Sub(wakeword, "wakeword")
}

// AuxiliaryFS returns a filesystem rooted at the auxiliary models directory.
func AuxiliaryFS() (fs.FS, error) {
	return fs.Sub(auxiliary, "auxiliary")
}

// ReadWakewordModel reads a wakeword model file by name.
func ReadWakewordModel(name string) ([]byte, error) {
	return wakeword.ReadFile(fmt.Sprintf("wakeword/%s", name))
}

// ReadAuxiliaryModel reads an auxiliary model file by name.
func ReadAuxiliaryModel(name string) ([]byte, error) {
	return auxiliary.ReadFile(fmt.Sprintf("auxiliary/%s", name))
}

// WakewordConfig reads the wakewords.yaml configuration file.
func WakewordConfig() ([]byte, error) {
	return wakeword.ReadFile("wakeword/wakewords.yaml")
}
