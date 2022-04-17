extern crate reqwest;

use std::{env, thread, time, process};

const ACCOUNT_ID: &str = "7481";
const ZONE_ID: &str = "cjbotta.ro";
const RECORD_ID: &str = "13635713";

fn main() {
    let mut ip_address: Option<String> = None;
    let token = env::var("TOKEN").expect("TOKEN not set");
    let sleep_for = time::Duration::from_secs(600); // 10 minutes

    loop {
        let new_ip_address = get_ip_address();

        if ip_address_changed(&ip_address, &new_ip_address) {
            match ip_address {
                None => println!("IP address changed: None -> {}", new_ip_address),
                Some(ip_address) => println!("IP address changed: {} -> {}", ip_address, new_ip_address)
            };
            update_dns(&token, &new_ip_address);
            ip_address = Some(new_ip_address);
        }

        thread::sleep(sleep_for);
    }
}

fn update_dns(token: &str, ip_address: &str) {
    use reqwest::StatusCode;

    let client = reqwest::blocking::Client::new();
    let url = format!("https://api.dnsimple.com/v2/{}/zones/{}/records/{}",
        ACCOUNT_ID, ZONE_ID, RECORD_ID
    );
    let body = format!("{{\"content\":\"{}\"}}", ip_address);
    let resp = client
        .patch(&url)
        .bearer_auth(token)
        .header("Content-Type", "application/json")
        .body(body)
        .send();

    match resp {
        Ok(resp) => match resp.status() {
            StatusCode::OK => (),
            status_code => {
                println!("API request failed with code {:?}", status_code);
                match resp.text() {
                    Ok(body) => println!("{}", body),
                    Err(reason) => println!("{:?}", reason)
                }
                process::exit(1);
            }
        },
        Err(reason) => panic!("{:?}", reason)
    }
}

fn ip_address_changed(old_ip_address: &Option<String>, ip_address: &str) -> bool {
    match old_ip_address {
        None => true,
        Some(old_ip_address) => old_ip_address != ip_address
    }
}

fn get_ip_address() -> String {
    let resp = reqwest::blocking::get("http://icanhazip.com/")
        .expect("Failed to get ip address");

    match resp.text() {
        Ok(s) => String::from(s.trim()),
        Err(reason) => panic!("{:?}", reason)
    }
}