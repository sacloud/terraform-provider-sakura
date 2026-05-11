// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package ftps

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jlaffaye/ftp"
)

// UploadFile uploads the given local file to the SAKURA Cloud FTPS server
// using explicit TLS. The filename on the server side is derived from
// filepath.Base(file).
func UploadFile(ctx context.Context, user, pass, host, file string) error {
	f, err := os.Open(filepath.Clean(file))
	if err != nil {
		return fmt.Errorf("opening file[%s] failed: %s", file, err)
	}
	defer f.Close() //nolint

	compCh := make(chan struct{})
	errCh := make(chan error)

	go func() {
		defer close(compCh)
		defer close(errCh)

		conn, err := ftp.Dial(
			fmt.Sprintf("%s:%d", host, 21),
			ftp.DialWithTimeout(30*time.Minute),
			ftp.DialWithExplicitTLS(&tls.Config{
				ServerName: host,
				MinVersion: tls.VersionTLS12,
				MaxVersion: tls.VersionTLS13,
			}))
		if err != nil {
			errCh <- fmt.Errorf("failed to connect to FTP server[%s]: %w", host, err)
			return
		}
		defer conn.Quit() //nolint:errcheck

		if err := conn.Login(user, pass); err != nil {
			errCh <- fmt.Errorf("failed to login to FTP server[%s]: %w", host, err)
			return
		}

		if err := conn.Stor(filepath.Base(file), f); err != nil {
			errCh <- fmt.Errorf("failed to upload file[%s]: %w", host, err)
			return
		}

		compCh <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		f.Close() //nolint
		return ctx.Err()
	case err := <-errCh:
		return err
	case <-compCh:
		return nil
	}
}
