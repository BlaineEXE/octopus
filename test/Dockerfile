FROM opensuse/leap:15.0
# FROM ubuntu:bionic

# Test hosts must run an ssh server daemon
RUN zypper --gpg-auto-import-keys --non-interactive install \
        net-tools \
        openssh
# RUN apt-get update -y && \
#     apt-get install -y net-tools openssh-server openssh-sftp-server openssh-client

COPY .ssh/* /root/.ssh/
COPY .ssh/id_rsa /etc/ssh/ssh_host_rsa_key
RUN chmod 600 /root/.ssh/id_rsa /root/.ssh/authorized_keys /etc/ssh/ssh_host_rsa_key

RUN mkdir -p /var/run/sshd

CMD [ "/usr/sbin/sshd", "-D", "-e" ]

# The integration tests we'll run from the test container as 'tester' user
RUN useradd --create-home tester
WORKDIR /home/tester

COPY .ssh/* /home/tester/.ssh/
RUN chmod 600 /home/tester/.ssh/id_rsa /home/tester/.ssh/authorized_keys

RUN mkdir /home/tester/tests
COPY tests/* /home/tester/tests/
RUN chown --recursive tester /home/tester


# Allow /etc/octopus to be read/writeable by our 'tester' user
RUN mkdir -p /etc/octopus && chmod --recursive 777 /etc/octopus
# The binary under test should be copied to the tests/_output dir
COPY _output/octopus /usr/bin
