# Systemd Unit

If you are using distribution packages or the copr repository, you don't need to deal with these files!

Copy freeswitch_exporter to directory `/usr/sbin/`

The unit file in this directory is to be put into `/etc/systemd/system`.

```shell
systemctl enable freeswitch_exporter.service
systemctl start freeswitch_exporter.service
```
