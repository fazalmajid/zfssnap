# zfssnap
Cron job to manage ZFS snapshots

> Note: do NOT use this on Ubuntu or any similar ZFS-on-Linux distributions that have the `zsysd` daemon for integration with `systemd` and other management tasks. It seems `zsysd` takes objection to some of the snapshots created by `zvault` and will crash, then be restarted by `systemd`, and so on *ad nauseam*, eating up half your CPU in the process. Alpine Linux does not seem to have this problem.

zfssnap is a small utility written in Go to manage automatic ZFS snapshots on
a system that supports it, like Illumos (Solaris), Linux or FreeBSD. I have
only tested it on SmartOS and Linux, so your mileage may vary.

It recursively creates snapshots on the ZFS filesystems supplied to it on the
command-line. It will keep 14 days of daily snapshots and 24 hourly snapshots.

My cron job for it is:

    0 * * * * /usr/bin/env TZ=US/Pacific /opt/root/bin/zfssnap zones/majid

The resulting snapshots are:

    [root@emurlahn ~]# date
    August  4, 2015 07:38:55 PM UTC
    [root@emurlahn ~]# zfs list -t snapshot -r zones/majid
    NAME                           USED  AVAIL  REFER  MOUNTPOINT
    zones/majid@daily-2015-07-22   386M      -   454G  -
    zones/majid@daily-2015-07-23   313M      -   454G  -
    zones/majid@daily-2015-07-24  41.7M      -   454G  -
    zones/majid@daily-2015-07-25  8.26M      -   454G  -
    zones/majid@daily-2015-07-26   135M      -   454G  -
    zones/majid@daily-2015-07-27   191M      -   454G  -
    zones/majid@daily-2015-07-28   192M      -   454G  -
    zones/majid@daily-2015-07-29   195M      -   454G  -
    zones/majid@daily-2015-07-30   300M      -   454G  -
    zones/majid@daily-2015-07-31   192M      -   454G  -
    zones/majid@daily-2015-08-01   111M      -   454G  -
    zones/majid@daily-2015-08-02  36.2M      -   454G  -
    zones/majid@daily-2015-08-03  47.2M      -   454G  -
    zones/majid@hourly-13         18.5M      -   454G  -
    zones/majid@hourly-14         12.4M      -   454G  -
    zones/majid@hourly-15         6.93M      -   454G  -
    zones/majid@hourly-16         3.45M      -   454G  -
    zones/majid@hourly-17         4.17M      -   454G  -
    zones/majid@hourly-18         5.00M      -   454G  -
    zones/majid@hourly-19         3.19M      -   454G  -
    zones/majid@hourly-20         7.61M      -   454G  -
    zones/majid@hourly-21          476K      -   454G  -
    zones/majid@hourly-22          479K      -   454G  -
    zones/majid@hourly-23         5.83M      -   454G  -
    zones/majid@hourly-00             0      -   454G  -
    zones/majid@daily-2015-08-04      0      -   454G  -
    zones/majid@hourly-01         7.11M      -   454G  -
    zones/majid@hourly-02         6.93M      -   454G  -
    zones/majid@hourly-03          510K      -   454G  -
    zones/majid@hourly-04          494K      -   454G  -
    zones/majid@hourly-05         10.1M      -   454G  -
    zones/majid@hourly-06         13.7M      -   454G  -
    zones/majid@hourly-07         14.8M      -   454G  -
    zones/majid@hourly-08         15.6M      -   454G  -
    zones/majid@hourly-09         3.83M      -   454G  -
    zones/majid@hourly-10         4.20M      -   454G  -
    zones/majid@hourly-11         15.5M      -   454G  -
    zones/majid@hourly-12         13.4M      -   454G  -

To build it, simply run make, and copy the zfssnap binary to the location of
your choice.

# zfsvault
This program is meant for offline backup to a "vault" drive.

My own backup server is a SmartOS machine with 2x mirrored 14TB Seagate Exos
Enterprise 14 drives. Clients back up to it using `rsync`. The vault is a pair
of WD My Book 14TB drives in rotation. My current backup set is about 5TB (and
no longer fits on my previous My Passport 5TB USB drives) and it takes about
12 hours to do the first transfer, which is why it makes sense to rotate
drives more frequently than the 14 days it takes for the oldest snapshot that
would be on it to be deleted on the source by `zfssnap`, this way you only
need to back up up to 14 day's worth of activity rather than since the
beginning of time.

`zfsvault` is a small utility written in Go to sync automatic ZFS snapshots
created using `zfssnap` between pools on a system that supports it, like
Illumos (Solaris) or FreeBSD. I have only tested it on SmartOS, so your
mileage may vary.

It uses `zfs send` to send snapshots, incrementally if possible, between two
zpools. It will ignore zfssnap hourly snapshots as the name is ambiguous. The
filesystems received will have the property `canmount=off` set otherwise two
datasets with the same mountpoint would cause a conflict in case of reboot.

My cron job for it is:

    30 0 * * * /usr/bin/env TZ=UTC PATH=/usr/local/bin:/usr/sbin:/usr/bin \
    /root/bin/zfsvault -target vault -m zones > /var/log/zfsvault.log 2>&1

I would also recommend using `zfs diff` to see what files changed between
backups, as you may find you have redundant backups that could be rationalized
to save on disk space.
