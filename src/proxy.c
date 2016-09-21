/*
 * This file is part of cc-oci-runtime.
 *
 * Copyrighth (C) 2016 Intel Corporation
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License
 * as published by the Free Software Foundation; either version 2
 * of the License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
 */

#include <gio/gunixsocketaddress.h>
#include "proxy.h"

#define PROXY_SOCKET	LOCALSTATEDIR "/run/cc-oci-runtime/proxy.sock"

int cc_proxy_connect(struct cc_proxy *proxy)
{
	GSocketAddress *addr;
	GError *error = NULL;
	int ret = -1;

	addr = g_unix_socket_address_new(PROXY_SOCKET);
	if (!addr) {
		g_critical("socket path does not exist: %s", PROXY_SOCKET);
		goto out_addr;
	}

	proxy->socket = g_socket_new(G_SOCKET_FAMILY_UNIX,
				     G_SOCKET_TYPE_SEQPACKET,
				     G_SOCKET_PROTOCOL_DEFAULT, &error);
	if (!proxy->socket) {
		g_critical("failed to create socket: %s", error->message);
		g_error_free(error);
		goto out_socket;
	}

	ret = g_socket_connect(proxy->socket, addr, NULL, &error);
	if (!ret) {
		g_critical("failed to connect to socket: %s", error->message);
		g_error_free(error);
		goto out_connect;
	}

	g_debug("connected to proxy (" LOCALSTATEDIR ")");
	return ret;

out_connect:
	g_clear_object(&proxy->socket);
out_socket:
	g_object_unref(addr);
out_addr:
	return ret;
}

int cc_proxy_disconnect(struct cc_proxy *proxy)
{
	g_socket_close(proxy->socket, NULL);
	g_clear_object(&proxy->socket);
	return 0;
}
