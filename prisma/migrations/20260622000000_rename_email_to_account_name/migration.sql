-- Rename the login identifier: email -> account_name (immutable, lowercase).
ALTER TABLE "app_user" RENAME COLUMN "email" TO "account_name";
ALTER INDEX "app_user_email_key" RENAME TO "app_user_account_name_key";

-- Drop the vestigial, unused Invite.email column (link-based invites only).
ALTER TABLE "invite" DROP COLUMN "email";
