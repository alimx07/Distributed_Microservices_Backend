#!/usr/bin/env bash
set -euo pipefail

# STEPS NEEDE
# 1- install terrafrom
# 2- install terragrunt
# 3- install terragrunt config tool
# 4- install atlantis (cool name)
# 5- start atlantis as service (rerun)


# Ensure they are found
apt-get update -y
apt-get install -y curl unzip git jq



# terrafrom
curl -fsSL "https://releases.hashicorp.com/terraform/${tf_version}/terraform_${tf_version}_linux_amd64.zip" -o /tmp/terraform.zip
unzip -o /tmp/terraform.zip -d /usr/local/bin terraform
chmod +x /usr/local/bin/terraform
rm /tmp/terraform.zip


#terragrunt
curl -fsSL "https://github.com/gruntwork-io/terragrunt/releases/download/v${tg_version}/terragrunt_linux_amd64.zip" -o /tmp/terragrunt.zip
unzip -o /tmp/terragrunt.zip -d /usr/local/bin terragrunt_linux_amd64
chmod +x /usr/local/bin/terragrunt_linux_amd64
rm /tmp/terragrunt.zip


#terragrunt atlantis config tool
curl -fsSL "https://github.com/transcend-io/terragrunt-atlantis-config/releases/download/v${tgac_version}/terragrunt-atlantis-config_${tgac_version}_linux_amd64" \
  -o /usr/local/bin/terragrunt-atlantis-config
chmod +x /usr/local/bin/terragrunt-atlantis-config


#atlantis
curl -fsSL "https://github.com/runatlantis/atlantis/releases/download/v${atlantis_version}/atlantis_linux_amd64.zip" \
  -o /tmp/atlantis.zip
unzip -o /tmp/atlantis.zip -d /usr/local/bin atlantis
chmod +x /usr/local/bin/atlantis
rm /tmp/atlantis.zip

# create user for atlantis
useradd --system --shell /bin/bash --create-home atlantis
mkdir -p /var/lib/atlantis /etc/atlantis
chown atlantis:atlantis /var/lib/atlantis


# repos_yaml=$(cat /etc/atlantis/repos_yaml)
# repo
echo "${repos_yaml_b64}" | base64 -d > /etc/atlantis/repos.yaml
chown atlantis:atlantis /etc/atlantis/repos.yaml

# Systemd
cat > /etc/systemd/system/atlantis.service << 'EOF'
[Unit]
Description=Atlantis Terraform Pull Request Automation
After=network.target

[Service]
User=atlantis
WorkingDirectory=/var/lib/atlantis
EnvironmentFile=/etc/atlantis/env
ExecStart=/usr/local/bin/atlantis server \
  --repo-config=/etc/atlantis/repos.yaml \
  --data-dir=/var/lib/atlantis \
  --port=4141
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF


cat > /etc/atlantis/env << EOF
ATLANTIS_GH_USER=${gh_user}
ATLANTIS_GH_TOKEN=${gh_token}
ATLANTIS_GH_WEBHOOK_SECRET=${gh_webhook_secret}
ATLANTIS_REPO_ALLOWLIST=${repo_allowlist}
EOF
chmod 600 /etc/atlantis/env
chown atlantis:atlantis /etc/atlantis/env


# systemd
systemctl daemon-reload
systemctl enable atlantis
systemctl start atlantis