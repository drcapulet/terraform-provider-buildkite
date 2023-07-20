provider "buildkite" {}

resource "buildkite_team" "test" {
  name = "terraform_provider_test"

  privacy             = "VISIBLE"
  default_team        = false
  default_member_role = "MEMBER"
}

resource "buildkite_team_member" "member1" {
  role = "MEMBER"
	team_id = "VXfhnVUS78HavgtP55WhWGzT401guK38Vm9LMMeCgQD124m8xaKBRq0Fth=="
	user_id = "VXNbwSA9hwVPpMgUXu1dWIDWf45ZwU6J7deETygiLUrKBg2TZBxuDr6aKj=="
}

resource "buildkite_organization_settings" "devorg" {
    allowed_api_ip_addresses = ["10.1.100.33/24"]
}

resource "buildkite_pipeline" "repo2" {
  name       = "terraform_provider_buildkite_pipeline"
  repository = "git@github.com:org/repo2"
  steps      = file("./steps.yml")

  team {
    slug         = buildkite_team.test.slug
    access_level = "READ_ONLY"
  }

  deletion_protection = true
}

resource "buildkite_agent_token" "fleet" {
  description = "token used by build fleet"
}

resource "buildkite_cluster_queue" "queue1" {
  cluster_id = "Q2x1c3Rlci0tLTMzMDc0ZDhiLTM4MjctNDRkNC05YTQ3LTkwN2E2NWZjODViNg=="
  key = "dev"
  description = "Dev cluster queue"
}

resource "buildkite_cluster_agent_token" "token1" {
  cluster_id = "Q2x1c3Rlci0tLTMzMDc0ZDhiLTM4MjctNDRkNC05YTQ3LTkwN2E2NWZjODViNg==" 
  description = "agent token for Dev cluster" 
}

output "agent_token" {
  value = buildkite_agent_token.fleet.token
}

output "badge_url" {
  value = buildkite_pipeline.test.badge_url
}

