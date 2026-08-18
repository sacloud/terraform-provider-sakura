# Specify the ID in the format of {zone}/{id}: e.g. "tk1b/113801540562"
terraform import sakura_seg.foo '{zone}/{id}'

# You can also omit the zone
# Doing so implies the default zone specified in the provider configuration.
terraform import sakura_seg.foo '{id}'
