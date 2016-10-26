#!/bin/bash
#  This file is part of cc-oci-runtime.
#
#  Copyright (C) 2016 Intel Corporation
#
#  This program is free software; you can redistribute it and/or
#  modify it under the terms of the GNU General Public License
#  as published by the Free Software Foundation; either version 2
#  of the License, or (at your option) any later version.
#
#  This program is distributed in the hope that it will be useful,
#  but WITHOUT ANY WARRANTY; without even the implied warranty of
#  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
#  GNU General Public License for more details.
#
#  You should have received a copy of the GNU General Public License
#  along with this program; if not, write to the Free Software
#  Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.

# We run the travis script from an exploded `make dist` tarball to ensure
# `make dist` has the necessary files to compile and run the tests.
#
# Note: we don't use `make distcheck` here as we can't run everything we want
#       for discheck in travis.
chronic make dist
tarball=`ls -1 cc-oci-runtime-*.tar.xz`
chronic tar xvf "$tarball"
export tarball_dir=${tarball%.tar.xz}
