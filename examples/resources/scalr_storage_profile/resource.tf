resource "scalr_storage_profile" "example_google" {
  name    = "my-google-storage-profile"
  default = true
  google {
    storage_bucket = "my-bucket"
    encryption_key = "S5pst/kWvXUmpaIQ8kSb3mr+h4yrA+Q024mOMMO8Bog="
    project        = "playground"
    credentials    = <<EOF
    {
      "type": "service_account",
      "project_id": "playground",
      "private_key_id": "b185b5359...",
      "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
      "client_email": "sa@playground.iam.gserviceaccount.com",
      "client_id": "1234567890",
      "auth_uri": "https://accounts.google.com/o/oauth2/auth",
      "token_uri": "https://oauth2.googleapis.com/token",
      "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
      "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/sa%40playground.iam.gserviceaccount.com"
    }
    EOF
  }
}

resource "scalr_storage_profile" "example_azure" {
  name = "my-azure-storage-profile"
  azurerm {
    audience        = "awesome-audience"
    client_id       = "12345678-1234-1234-1234-123456789012"
    container_name  = "my-container"
    storage_account = "my-storage-account"
    tenant_id       = "12345678-1234-1234-1234-123456789012"
  }
}
