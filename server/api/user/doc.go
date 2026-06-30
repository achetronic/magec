// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

// Package user provides the Magec User-facing REST API.
//
// @title           Magec User API
// @version         1.0
// @description     User-facing API for Magec. Includes device pairing, voice proxies, and health checks.
//
// @host            localhost:8080
// @BasePath        /api/v1
//
// @schemes         http
//
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Enter your bearer token as: Bearer <token>
package user
