�ڵ�̨����( centos7.x )����,������redis��cfs-volmgr��cfs-metanode��cfs-datanode��cfs-client��cfs-fuseclient

��װ��
1���������IPΪ 10.8.65.94
2��yum ��װ mysql �� redis ������
3����ѹ cfs.tar.gz �� /home 
4��cd /home/cfs/ 
5��dos2unix install.sh
6��chmod +x install.sh
7��./install.sh

���ã�
1��cd /usr/local/conf/ 
2���ֱ����� cfs-volmgr.ini cfs-metanode.ini cfs-datanode.ini �����׶������ã�

������
service cfs-volmgr start
service cfs-metanode start
service cfs-datanode start

���� volume:

���� cfs-client.ini 
/home/cfs/cfs-client /home/cfs/cfs-client.ini createvol yourname 40

���� volume:
���� cfs-fuseclient.ini 
/usr/local/bin/cfs-fuseclient /usr/local/conf/cfs-fuseclient.ini
