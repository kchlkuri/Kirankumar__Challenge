#! /bin/bash
aws s3 cp s3://edshop-web-data/bootstrap.zip /tmp/
unzip /tmp/bootstrap.zip
yum install ansible -y
cd bootstrap
ansible-playbook setup.yaml