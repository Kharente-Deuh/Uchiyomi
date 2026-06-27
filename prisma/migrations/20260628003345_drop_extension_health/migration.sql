-- DropForeignKey
ALTER TABLE "extension_error_log" DROP CONSTRAINT "extension_error_log_extension_id_fkey";

-- DropTable
DROP TABLE "extension_error_log";

-- AlterTable
ALTER TABLE "extension" DROP COLUMN "health",
DROP COLUMN "consecutive_failures",
DROP COLUMN "last_error_at",
DROP COLUMN "last_error_message";

-- DropEnum
DROP TYPE "ExtensionHealth";
