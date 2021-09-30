# Steps Below
sudo adduser sec

sudo adduser sec sudo

sudo su - sec

sudo apt-get update

sudo add-apt-repository -y ppa:longsleep/golang-backports

sudo sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'

wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo apt-key add -

sudo apt-get update

sudo apt install -y software-properties-common
sudo apt install -y postgresql-13
sudo apt install -y python3
sudo apt install -y git
sudo apt install -y make
sudo apt install -y vim
sudo apt install -y tig
sudo apt install -y ncdu
sudo apt install -y htop
sudo apt install -y tmux
sudo apt install -y jq
sudo apt install -y curl
sudo apt install -y golang-go

# Could not create the dir without sudo which sets owner as root, so we chown to sec
sudo mkdir -p /mnt/sec/db

sudo chown -R sec /mnt/sec

/usr/lib/postgresql/13/bin/initdb -D /mnt/sec/db

sudo su - postgres

psql

CREATE ROLE sec WITH SUPERUSER LOGIN CREATEDB;

CREATE DATABASE sec OWNER sec;
