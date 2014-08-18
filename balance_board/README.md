f7u12 + Nintendo Wii Balance Board
==================================

The program in this directory enables input from a Wii Balance Board to
be fed to the f7u12 2048 game.  Since it relies on functionality provided
by the Linux kernel (hid-wiimote.ko) and BlueZ 5.21, it only works on
Linux at present.

Usage
=====

Install the BlueZ 5.x stack on a modern Linux kernel. BlueZ 5.21 and
Linux 3.16.1 have been tested to work.

```
$ sudo systemctl start bluetooth.service
$ bluetoothctl
[bluetooth]# power on
[bluetooth]# agent on
[bluetooth]# default-agent
[bluetooth]# discoverable on
[bluetooth]# scan on
# now press the sync button on the balance board
[bluetooth]# pair 00:24:44:DC:0B:25
[bluetooth]# connect 00:24:44:DC:0B:25
[bluetooth]# quit
$ sudo ./balance_board
```
