#!/bin/sh

# XXXHACK: One day we will convert it to Ansible playbook

# XXXHACK: This is considered a development machine (VPS). Production shouldn't have any of that.
# XXXHACK: Production machine will only contain the binary of `sec` and a config file.

# install add-apt-repository tool https://itsfoss.com/add-apt-repository-command-not-found/
apt-get install -y software-properties-common

# get the Go's PPA installed so we get the latest version of Go
apt-key adv --keyserver keyserver.ubuntu.com --recv-key C99B11DEB97541F0
add-apt-repository ppa:longsleep/golang-backports
add-apt-repository https://cli.github.com/packages

# Update the repos for packages and upgrade all packages
apt update -y
apt upgrade -y
apt autoremove -y

apt install -y \
	postgresql \
	golang-go git make vim tig ncdu htop tmux jq \
	postgresql ack curl gh

# we should figure out how to make a DB automatically too
