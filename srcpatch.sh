# !-- DO NOT USE THIS FOR DEVELOPMENT --!
# This is only to source patch files in production deployments, if prisma isn't configured already.
printf "//@ts-nocheck\n$(cat src/routes/backends/lookup.ts)" > src/routes/backends/lookup.ts
printf "//@ts-nocheck\n$(cat src/routes/forward/lookup.ts)" > src/routes/forward/lookup.ts
printf "//@ts-nocheck\n$(cat src/routes/user/lookup.ts)" > src/routes/user/lookup.ts
printf "//@ts-nocheck\n$(cat src/routes/getPermissions.ts)" > src/routes/getPermissions.ts
