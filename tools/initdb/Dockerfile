FROM oraclelinux:7-slim

ARG KEY=https://repo.mysql.com/RPM-GPG-KEY-mysql
ARG REPO=https://repo.mysql.com

ARG PACKAGE_URL_SHELL=$REPO/yum/mysql-tools-community/el/7/x86_64/mysql-shell-8.0.13-1.el7.x86_64.rpm

COPY ./migrations /migrations

# Install server
RUN rpmkeys --import $KEY \
  && yum install -y $PACKAGE_URL_SHELL libpwquality