FROM opensuse/leap:15.0

# Test hosts must run an ssh server daemon
RUN zypper --gpg-auto-import-keys --non-interactive install \
        net-tools \
        openssh

WORKDIR /root

COPY .ssh/* /root/.ssh/
COPY .ssh/id_rsa /etc/ssh/ssh_host_rsa_key

CMD [ "/usr/sbin/sshd", "-D" ]

# The integration tests we'll run from the test container
RUN mkdir /root/tests
COPY tests/* /root/tests/

# The binary under test should be copied to the tests/_output dir
COPY _output/octopus /usr/bin
