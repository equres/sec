#!/bin/sh

set -ex

# ANSIBLE_OPTS=
# if [ "$1" = "-v" ]; then
# 	ANSIBLE_OPTS=-vvvv # verbose
# fi

SERVER=$1
HOST=
USER=

if [ $SERVER = "dev" ]; then
	HOST="10.7.7.7"
	USER="vagrant"
elif [ $SERVER = "prod" ]; then
	HOST="192.99.161.20"
	USER="ubuntu"
fi


ssh $USER@$HOST sudo apt-get install -y python3
ansible ${ANSIBLE_OPTS} -u $USER -i ${HOST}, all -m ping
ansible-playbook --become-user="$USER" -i hosts.yml  -e "ansible_python_interpreter=/usr/bin/python3 host_names=$HOST ssh_user=$USER" playbook.yml