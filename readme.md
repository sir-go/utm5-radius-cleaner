### UTM5 RADIUS sessions cleaner

1. get active sessions list `rpcf_radius_get_active_sessions`
2. kill sessions with `traf_last_update_date` > `%session_lifetime%` (> 10 min) - `rpcf_radius_drop_session`
3. if failed then restart utm_radius and retry
4. if IP == 0.0.0.0 then kill session on the ISG-Router via telnet `clear sss session user <ip>`

---
conf:
```
[service]
    location    = "Europe/Moscow"

[billing]
    api_url = "https://utm.ttnet.ru/api/v1"
    username = ""
    password = ""
    session_life_time = "10m"
    ssh_username = "<%USERNAME%>"
    ssh_password = "<%PASSWORD%>"
    ssh_timeout = "1m"
    radius_restart_cmd = "/etc/init.d/utm5_radius stop && /etc/init.d/utm5_radius start"

    [billing.hosts]
        tih = ""
        kor = ""

[cisco]
    address = ""
    username = "<%USERNAME%>"
    password = "<%PASSWORD%>"
    drop_session_command = "clear sss session user %s"
    telnet_cmd_timeout = "10s"
```
