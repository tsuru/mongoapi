# -*- coding: utf-8 -*-

# Copyright 2013 mongoapi authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

import os
from fabric.api import env, local, put, run

current_dir = os.path.abspath(os.path.dirname(__file__))
env.user = 'ubuntu'
env.bin_path = '/home/%s/bin' % env.user


def build():
    local("go clean")
    local("GOOS=linux GOARCH=amd64 go build")


def send():
    run("mkdir -p %s" % env.bin_path)
    put(os.path.join(current_dir, "mongoapi"), env.bin_path)


def stop():
    run("circusctl stop mongoapi")


def start():
    run("circusctl start mongoapi")


def clean():
    local("rm mongoapi")


def deploy():
    build()
    stop()
    send()
    start()
    clean()
