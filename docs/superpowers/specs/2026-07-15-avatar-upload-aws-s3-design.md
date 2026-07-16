# Avatar Upload to AWS S3 Design

## Goal

Make employee avatar uploads use AWS S3 exclusively. Store each avatar under
`hrm-app/avatars/<uuid>.<ext>` in the configured bucket and persist its public
AWS URL on the employee record.

## Scope

This change covers the shared storage adapter and employee avatar flow. Existing
avatar API routes, multipart field names, authorization, size limits, and MIME
validation remain unchanged. Skill icons, leave attachments, and contract
attachments continue using the same shared adapter, so they move from the old
Supabase-specific transport to AWS S3 without changing their object prefixes.

Supabase support, configuration, URL construction, and documentation are removed.
No generic upload endpoint is added.

## Storage Configuration

`StorageConfig` contains four required values:

- `STORAGE_ACCESS_KEY`
- `STORAGE_SECRET_KEY`
- `STORAGE_REGION`
- `STORAGE_BUCKET`

`STORAGE_ENDPOINT` is removed because the AWS SDK resolves the regional S3
endpoint from `STORAGE_REGION`. Storage configuration fails loudly during boot
when any required value is missing. This prevents a server that starts normally
but fails only when the first upload occurs.

The S3 client uses static credentials, configured region, AWS default endpoint
resolution, and virtual-hosted addressing. It does not set a custom base endpoint
or path-style addressing.

## Object Keys and Public URLs

Avatar keys use:

```text
hrm-app/avatars/<uuid>.<lowercase-extension>
```

The storage adapter continues generating the UUID. The employee service passes
`hrm-app/avatars` as the subdirectory.

Public URLs use:

```text
https://<bucket>.s3.<region>.amazonaws.com/<key>
```

This design assumes the configured bucket permits public reads for uploaded
objects. Upload credentials remain server-side and are never returned.

## Avatar Request Flow

1. Authenticated client sends `PATCH /api/v1/employees/me/avatar`, or an
   authorized administrator sends `PATCH /api/v1/employees/:id/avatar`.
2. Handler reads multipart field `avatar`, enforces the 5 MB limit, and accepts
   PNG, JPEG, or WEBP.
3. Employee service sniffs actual bytes and rejects spoofed content types.
4. Storage adapter uploads the bytes to `hrm-app/avatars/<uuid>.<ext>`.
5. Employee repository stores the returned AWS public URL.
6. After the database update succeeds, the previous avatar is deleted on a
   best-effort basis.

If upload fails, the employee record remains unchanged. If the database update
fails after upload, the service best-effort deletes the newly uploaded object to
avoid an orphan. Failure to delete an old avatar does not roll back a successful
avatar replacement.

## Error Handling

- Missing storage configuration: server boot fails with named missing variables.
- Invalid or oversized image: existing HTTP 400 behavior remains.
- AWS upload failure: request fails and employee row remains unchanged.
- Employee persistence failure: request fails and new object is cleaned up when
  possible.
- Old-object deletion failure: request still succeeds because the new URL is
  already persisted.

## Tests and Verification

Tests encode these requirements:

- AWS public URL contains configured bucket, region, and object key.
- Uploaded avatar subdirectory is exactly `hrm-app/avatars`.
- Valid images reach storage; spoofed or oversized files do not.
- Failed employee persistence deletes the newly uploaded object.
- Successful replacement deletes the previous avatar only after persistence.
- Missing AWS storage configuration is rejected.
- No Supabase-specific URL or endpoint behavior remains.

Verification runs the focused unit/service tests, `go test ./...`, and a temporary
live S3 probe. The probe uploads a tiny valid PNG under `hrm-app/avatars`, verifies
the object through S3, then deletes it. It does not modify an employee record.

## Success Criteria

- Avatar object appears at `hrm-app/avatars/<uuid>.<ext>` in the configured AWS
  bucket.
- Returned URL uses the AWS S3 public URL format and is persisted for the employee.
- Existing avatar validation and authorization behavior stays intact.
- Supabase-specific code and documented configuration are gone.
- Automated verification completes without hidden skips; any unavailable DB-backed
  verification is reported explicitly.
