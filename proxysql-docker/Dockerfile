FROM ubuntu:14.04

RUN apt-get update && apt-get install -y \
    automake \
    cmake \
    make \
    g++ \
    gcc \
    gdb \
    gdbserver \
    git \
    libmysqlclient-dev \
    libssl-dev \
    libtool \
    mysql-client \
  && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists/

RUN cd /opt && \
    git clone -b v1.2.2 https://github.com/sysown/proxysql.git && \
    cd /opt/proxysql && \
    make clean && make > /tmp/proxysql_compile.log

ADD proxysql.cnf /etc/proxysql.cnf

COPY proxysql-entry.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

VOLUME /var/lib/proxysql

COPY proxysql_galera_checker.sh /opt/proxysql/proxysql_galera_checker.sh
RUN chmod +x /opt/proxysql/proxysql_galera_checker.sh

EXPOSE 3306 6032

CMD [""]
