#!/bin/bash

IPTABLES_ZIP_URL=http://www.netfilter.org/projects/iptables/files/
IPTABLES_ZIP_FILE_NAME=iptables-1.8.2

apt-get update
apt-get install -y ipset
apt-get install -y wget
apt-get install -y make
apt-get install -y gcc

wget ${IPTABLES_ZIP_URL}${IPTABLES_ZIP_FILE_NAME}.tar.bz2
tar xjf ${IPTABLES_ZIP_FILE_NAME}.tar.bz2 && cd ${IPTABLES_ZIP_FILE_NAME}
./configure --prefix=/usr  \
            --sbindir=/sbin    \
            --disable-nftables \
            --enable-libipq    \
            --with-xtlibdir=/lib/xtables && make

make install < /dev/null
ln -sfv ../../sbin/xtables-legacy-multi /usr/bin/iptables-xml
for file in ip4tc ip6tc ipq iptc xtables
do
  mv -v /usr/lib/lib${file}.so.* /lib &&
  ln -sfv ../../lib/$(readlink /usr/lib/lib${file}.so) /usr/lib/lib${file}.so
done

rm -rf ../${IPTABLES_ZIP_FILE_NAME}*

apt-get purge -y wget
apt-get purge -y make
apt-get purge -y gcc