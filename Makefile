SHELL := /bin/bash
TARGETS = proxima

# http://docs.travis-ci.com/user/languages/go/#Default-Test-Script
fmt:
	go fmt ./...

imports:
	goimports -w .

all: fmt test
	go build

install:
	go install

clean:
	go clean
	rm -fv coverage.out
	rm -fv $(TARGETS)
	rm -fv *.x86_64.rpm
	rm -fv debian/proxima*.deb
	rm -rfv debian/proxima/usr

cover:
	go get -d && go test -v	-coverprofile=coverage.out
	go tool cover -html=coverage.out

proxima: imports
	go build cmd/proxima/proxima.go

deb: $(TARGETS)
	mkdir -p debian/proxima/usr/sbin
	cp $(TARGETS) debian/proxima/usr/sbin
	cd debian && fakeroot dpkg-deb --build proxima .

# rpm building via vagrant
SSHCMD = ssh -o StrictHostKeyChecking=no -i vagrant.key vagrant@127.0.0.1 -p 2222
SCPCMD = scp -o port=2222 -o StrictHostKeyChecking=no -i vagrant.key

rpm: $(TARGETS)
	mkdir -p $(HOME)/rpmbuild/{BUILD,SOURCES,SPECS,RPMS}
	cp ./packaging/proxima.spec $(HOME)/rpmbuild/SPECS
	cp $(TARGETS) $(HOME)/rpmbuild/BUILD
	./packaging/buildrpm.sh proxima
	cp $(HOME)/rpmbuild/RPMS/x86_64/proxima*rpm .

# Helper to build RPM on a RHEL6 VM, to link against glibc 2.12
vagrant.key:
	curl -sL "https://raw.githubusercontent.com/mitchellh/vagrant/master/keys/vagrant" > vagrant.key
	chmod 0600 vagrant.key

# Don't forget to vagrant up :) - and add your public key to the guests authorized_keys
setup: vagrant.key
	$(SSHCMD) "sudo yum install -y sudo yum install http://ftp.riken.jp/Linux/fedora/epel/6/i386/epel-release-6-8.noarch.rpm"
	$(SSHCMD) "sudo yum install -y golang git rpm-build"
	$(SSHCMD) "mkdir -p /home/vagrant/src/github.com/miku"
	$(SSHCMD) "cd /home/vagrant/src/github.com/miku && git clone https://github.com/miku/proxima.git"

rpm-compatible: vagrant.key
	$(SSHCMD) "cd /home/vagrant/src/github.com/miku/proxima && GOPATH=/home/vagrant go get ./..."
	$(SSHCMD) "cd /home/vagrant/src/github.com/miku/proxima && git pull origin master && pwd && GOPATH=/home/vagrant make clean rpm"
	$(SCPCMD) vagrant@127.0.0.1:/home/vagrant/src/github.com/miku/proxima/*rpm .

# local rpm publishing
REPOPATH = /usr/share/nginx/html/repo/CentOS/6/x86_64

publish: rpm-compatible
	cp proxima-*.rpm $(REPOPATH)
	createrepo $(REPOPATH)
