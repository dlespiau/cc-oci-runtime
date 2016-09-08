/*
 *  This file is part of cc-shim.
 *
 *  Copyright (C) 2016 Intel Corporation
 *
 *  cc-shim is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 2 of the License, or
 *  (at your option) any later version.
 *
 *  cc-shim is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with cc-shim.  If not, see <http://www.gnu.org/licenses/>.
 */

#include <fcntl.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/socket.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <unistd.h>

#if 0
typedef struct {
	int ctl_pipe, io_pipe;
} shim_t;

static int fd_set_blocking(int fd, bool blocking)
{
	int flags, s;

	flags = fcntl(fd, F_GETFL, 0);
	if (flags == -1)
		return -1;

	if (blocking)
		flags &= ~O_NONBLOCK;
	else
		flags |= O_NONBLOCK;

	s = fcntl(fd, F_SETFL, flags);
	if (s == -1)
		return -1;

	return 0;
}

static shim_t shim;

int main(int argc, char *argv[])
{
	int ret;
	const char *fifo_path;

	if (argc != 2) {
		fprintf(stderr, "Usage: %s fifo\n", argv[0]);
		exit(1);
	}

	fifo_path = argv[1];

#if 0
	ret = mkfifo(TEST_FIFO, O_RDWR);
	if (ret != 0) {
		perror("mkfifo");
		ret = 1;
		goto err_mkfifo;
	}

	shim.ctl_pipe = open(fifo_path, O_RDWR, O_NONBLOCK);
	if (shim.ctl_pipe == -1) {
		perror("open ctl_pipe");
		ret = 1;
		goto err_open_fifo;
	}
#endif

	sock = socket(AF_UNIX, SOCK_STREAM, 0);

sock.connect(“MyBindName”)

# Wait for message
msg = sock.recv(100)
print msg

# Send reply
sock.send(“Hi there!n”)

# Block until new message arrives
msg = sock.recv(100)

# When the socket is closed cleanly, recv unblocks and returns “”
if not msg:
print “It seems the other side has closed its connection”

# Close it
sock.close()

	ret = fd_set_blocking(shim.ctl_pipe, false);
	if (ret == -1)
		goto err_non_blocking;

	fd_set rfds, wfds;
	struct timeval tv;
	int retval;

	FD_ZERO(&rfds);
	FD_SET(shim.ctl_pipe, &rfds);

	FD_ZERO(&wfds);
	FD_SET(shim.ctl_pipe, &wfds);
	/* Wait up to five seconds. */

	tv.tv_sec = 5;
	tv.tv_usec = 0;

	retval = select(1, &rfds, NULL, NULL, &tv);
	/* Don't rely on the value of tv now! */

	if (retval == -1)
		perror("select()");
	else if (retval) {
		printf("Data is available now.\n");
		if (FD_ISSET(shim.ctl_pipe, &rfds)) {
			printf("Data is available for reading.\n");
		} else if (FD_ISSET(shim.ctl_pipe, &wfds)) {
			printf("Data is available for writing.\n");
		}
	} else
		printf("No data within five seconds.\n");

	exit(EXIT_SUCCESS);

err_non_blocking:
err_open_fifo:
#if 0
err_mkfifo:
#endif
	return ret;
}

#endif

int main(void) { return 0; }
