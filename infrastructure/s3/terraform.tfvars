bucket_name = "cstore-{{CONTEXT}}"

role_users = [
"{{USER_ROLE}}/{{EMAIL_ADDRESS}}",
]

roles = [
"{{CONTAINER_ROLE}}",
]

users = [
  "{{IAM_USER}}",
]

tag_team          = "{{TEAM}}"
tag_contact-email = "{{EMAIL_ADDRESS}}"
tag_application   = "{{CONTEXT}}"
tag_environment   = "{{ENV}}"
tag_customer      = "{{CONTEXT}}"