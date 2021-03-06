# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.box = "flynn-base"
  config.vm.box_url = "https://dl.flynn.io/vagrant/flynn-base.json"
  config.vm.box_version = "> 0"

  config.vm.network "private_network", ip: "192.168.96.48"

  config.vm.provision "shell", privileged: false, inline: <<SCRIPT
    grep '^export GOPATH' ~/.bashrc || echo export GOPATH=~/go >> ~/.bashrc
    grep '^export PATH' ~/.bashrc || echo export PATH=\$PATH:~/go/bin:/vagrant/script >> ~/.bashrc
    GOPATH=~/go go get github.com/tools/godep

    # For controller tests
    sudo apt-get update
    sudo apt-get install -y postgresql postgresql-contrib
    sudo -u postgres createuser --superuser vagrant
    grep '^export PGHOST' ~/.bashrc || echo export PGHOST=/var/run/postgresql >> ~/.bashrc

    mkdir -p ~/go/src/github.com/flynn
    ln -s /vagrant ~/go/src/github.com/flynn/flynn
    grep ^cd ~/.bashrc || echo cd ~/go/src/github.com/flynn/flynn >> ~/.bashrc
SCRIPT
end
