// Copyright 2024, Asher Wrobel
/*
This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
use actix_web::{web, Responder};
use crate::verify_peer::verify_peer;

pub fn add_peer(ip: web::Path<String>) -> impl Responder {
    // Validate the IP address
    if !verify_peer(&ip) {
        return format!("Invalid peer: {}", ip);
    }
    // Add the peer to the list of peers in peers.txt
    std::fs::write("peers.txt", ip.to_string() + "\n").unwrap();
    println!("Peer added: {}", ip);
    format!("Peer added: {}", ip)
}
