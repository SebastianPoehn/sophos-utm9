# Copyright 2016 Vincent Landgraf. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

restd: *.go ../confd/*.go
	go build

run: restd
	./restd

test:
	go test -v -coverprofile=coverage.out
	go tool cover -html=coverage.out
