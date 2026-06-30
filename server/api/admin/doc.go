// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

// Package admin provides the Magec Admin REST API.
//
// @title           Magec Admin API
// @version         2.0
// @description     Administration API for managing Magec resources.
// @description     When server.adminPassword is configured, all /api/ endpoints require a Bearer token.
//
// @host            localhost:8081
// @BasePath        /api/v1/admin
//
// @schemes         http
//
// @securityDefinitions.apikey AdminAuth
// @in header
// @name Authorization
// @description Enter "Bearer {adminPassword}" to authenticate. Only required when server.adminPassword is set.
package admin
