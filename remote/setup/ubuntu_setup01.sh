# preps the EC2 Ubuntu Linux machine for use with local PostgreSQL

export LC_ALL=en_US.UTF-8

add-apt-repository --yes universe
apt update
apt --yes -o Dpkg::Options::="--force-confnew" upgrade

# allow SSH, HTTP and HTTPS traffic (still need to tweak SG settings in AWS)
ufw allow 22
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

apt --yes install locales-all
apt --yes install fail2ban

# Install Caddy using guide for Ubuntu: https://caddyserver.com/docs/install#debian-ubuntu-raspbian
apt --yes install -y debian-keyring debian-archive-keyring apt-transport-https
curl -L https://dl.cloudsmith.io/public/caddy/stable/gpg.key | sudo apt-key add -
curl -L https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt | sudo tee -a /etc/apt/sources.list.d/caddy-stable.list
apt update
apt --yes install caddy

# install postgres
sudo apt --yes install postgresql

# install migrate CLI tool
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate.linux-amd64 /usr/local/bin/migrate

echo "Script complete! Rebooting..."
reboot