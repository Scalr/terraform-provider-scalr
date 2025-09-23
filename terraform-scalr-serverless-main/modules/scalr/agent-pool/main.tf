terraform {
  required_providers {
    scalr = {
      source  = "scalr/scalr"
    }
  }
}

provider "scalr" {
  hostname = var.scalr_hostname
  token    = var.scalr_token_sub
}

data "http" "scalr_allowlist" {
  url = "https://main.scalr.dev/.well-known/allowlist.txt"
}

resource "scalr_agent_pool" "webhook" {
  name = var.agent_pool_name
}

resource "scalr_agent_pool_token" "webhook" {
  agent_pool_id = scalr_agent_pool.webhook.id
  description   = "Token for Scalr webhook agent"
}

locals {
  scalr_url = "https://${var.scalr_hostname}"
  # Split the response by newlines and filter out empty lines
  scalr_ips = [for ip in split("\n", trimspace(data.http.scalr_allowlist.response_body)) : "${ip}/32" if ip != ""]
}
