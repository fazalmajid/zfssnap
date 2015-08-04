# zfssnap
Cron job to manage ZFS snapshots

zfssnap is a small utility written in Go to manage automatic ZFS snapshots on
a system that supports it, like Illumos (Solaris) or FreeBSD. I have only
tested it on SmartOS, so your mileage may vary.

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
