#!/bin/bash

# This script is a hacky helper that should be removed once we add support
# for the TD oauth flow (acquiring initial token and refreshing)
#
# Usage:
# TD_CONSUMER_KEY=<your_application_id> ./get_td_token.sh

# DEPENDENCIES:
#   - python3
#   - curl
#   - jq

cd $(dirname $0)

[[ -n "$TD_CONSUMER_KEY" ]] || { echo "Please set TD_CONSUMER_KEY"; exit 1; }

TD_REDIRECT_URI=${TD_REDIRECT_URI:-https://127.0.0.1:42068/callback}

uriencode() {
    # cheating... i know...
    python -c "import urllib.parse; print(urllib.parse.quote('''$1'''))"
}

uridecode() {
    python -c "import urllib.parse; print(urllib.parse.unquote('''$1'''))"
}

get_new_auth() {
    xdg-open "https://auth.tdameritrade.com/auth?response_type=code&redirect_uri=$(uriencode $TD_REDIRECT_URI)&client_id=$(uriencode $TD_CONSUMER_KEY)%40AMER.OAUTHAP&scope=AccountAccess"
    echo "After authorizing the app, grab the \`code\` parameter and slam it in here:"
    echo -n " -> "
    read tdcode
    curl --silent \
         --request POST \
         --header "Content-Type: application/x-www-form-urlencoded" \
         --data "grant_type=authorization_code&refresh_token=&access_type=offline&code=${tdcode}&client_id=${TD_CONSUMER_KEY}&redirect_uri=$(uriencode $TD_REDIRECT_URI)" \
         "https://api.tdameritrade.com/v1/oauth2/token" > td_auth_refresh_info.json
}

refresh_auth() {
    refresh_token=$(uriencode $(jq -r '.refresh_token' td_auth_refresh_info.json))
    curl --silent \
         --request POST \
         --header "Content-Type: application/x-www-form-urlencoded" \
         --data "grant_type=refresh_token&refresh_token=${refresh_token}&access_type=&code=&client_id=${TD_CONSUMER_KEY}&redirect_uri=" \
         "https://api.tdameritrade.com/v1/oauth2/token" > td_auth_info.json
    access_token=$(jq -j '.access_token' td_auth_info.json)
    echo "Authorization info stored in td_auth_info.json"
    echo "Your access token is: $access_token"
}

[[ -r td_auth_refresh_info.json ]] || get_new_auth
refresh_auth
