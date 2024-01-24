tell application "iTerm2"
    activate
    select first window
    
    tell current session of current window
        write text "make start_db"
        write text "make start_redis"
         write text "sleep 5"
        split vertically with default profile
        split vertically with default profile
        split vertically with default profile
    end tell
    
    tell first session of current tab of current window
        write text "cd ~/go/src/verancial"
        write text "make start_login_local"
    end tell

    tell second session of current tab of current window
        write text "cd ~/go/src/verancial"
        write text "make start_api_local"
    end tell

    tell third session of current tab of current window
        write text "cd ~/go/src/verancial"
        write text "make start_app_integration_worker_local"
    end tell

    tell fourth session of current tab of current window
        write text "cd ~/go/src/verancial"
        write text "make start_report_process_worker_local"
    end tell

end tell