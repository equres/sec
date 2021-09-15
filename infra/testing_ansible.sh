#!/bin/sh

set -ex

ANSIBLE_OPTS=
if [ "$1" = "-v" ]; then
	ANSIBLE_OPTS=-vvvv # verbose
fi

HOST=192.99.161.20

ssh root@$HOST apt-get install -y python
ansible ${ANSIBLE_OPTS} -u root -i ${HOST}, all -m ping
ansible-playbook -i ${HOST}, playbook.yml