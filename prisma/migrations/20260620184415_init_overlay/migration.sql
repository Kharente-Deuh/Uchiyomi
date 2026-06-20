-- CreateEnum
CREATE TYPE "Role" AS ENUM ('ADMIN', 'USER');

-- CreateEnum
CREATE TYPE "UserStatus" AS ENUM ('ACTIVE', 'DISABLED');

-- CreateEnum
CREATE TYPE "AuthProvider" AS ENUM ('LOCAL', 'OIDC');

-- CreateEnum
CREATE TYPE "SeriesType" AS ENUM ('MANGA', 'WEBTOON', 'COMIC');

-- CreateEnum
CREATE TYPE "SeriesStatus" AS ENUM ('ONGOING', 'COMPLETED', 'HIATUS', 'UNKNOWN');

-- CreateEnum
CREATE TYPE "LibraryStatus" AS ENUM ('READING', 'STOPPED');

-- CreateEnum
CREATE TYPE "ReadingMode" AS ENUM ('PAGED', 'LONG_STRIP');

-- CreateEnum
CREATE TYPE "ReadingDirection" AS ENUM ('LTR', 'RTL');

-- CreateEnum
CREATE TYPE "DownloadStatus" AS ENUM ('PENDING', 'RUNNING', 'DONE', 'FAILED');

-- CreateEnum
CREATE TYPE "DownloadPriority" AS ENUM ('NEW_CHAPTER', 'BACKFILL');

-- CreateEnum
CREATE TYPE "ExtensionHealth" AS ENUM ('OK', 'ERROR');

-- CreateTable
CREATE TABLE "app_user" (
    "id" TEXT NOT NULL,
    "email" TEXT NOT NULL,
    "display_name" TEXT NOT NULL,
    "role" "Role" NOT NULL DEFAULT 'USER',
    "status" "UserStatus" NOT NULL DEFAULT 'ACTIVE',
    "can_manage_extensions" BOOLEAN NOT NULL DEFAULT false,
    "can_download" BOOLEAN NOT NULL DEFAULT false,
    "allow_nsfw" BOOLEAN NOT NULL DEFAULT false,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "app_user_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "auth_identity" (
    "id" TEXT NOT NULL,
    "user_id" TEXT NOT NULL,
    "provider" "AuthProvider" NOT NULL,
    "subject" TEXT NOT NULL,
    "password_hash" TEXT,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "auth_identity_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "session" (
    "id" TEXT NOT NULL,
    "user_id" TEXT NOT NULL,
    "user_agent" TEXT,
    "ip" TEXT,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "expires_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "session_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "invite" (
    "id" TEXT NOT NULL,
    "code" TEXT NOT NULL,
    "email" TEXT,
    "role" "Role" NOT NULL DEFAULT 'USER',
    "created_by_user_id" TEXT,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "expires_at" TIMESTAMP(3),
    "accepted_at" TIMESTAMP(3),

    CONSTRAINT "invite_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "subscription" (
    "id" TEXT NOT NULL,
    "user_id" TEXT NOT NULL,
    "series_id" TEXT NOT NULL,
    "status" "LibraryStatus" NOT NULL DEFAULT 'READING',
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "subscription_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "category" (
    "id" TEXT NOT NULL,
    "user_id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "order" INTEGER NOT NULL DEFAULT 0,

    CONSTRAINT "category_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "subscription_category" (
    "subscription_id" TEXT NOT NULL,
    "category_id" TEXT NOT NULL,

    CONSTRAINT "subscription_category_pkey" PRIMARY KEY ("subscription_id","category_id")
);

-- CreateTable
CREATE TABLE "reading_progress" (
    "id" TEXT NOT NULL,
    "user_id" TEXT NOT NULL,
    "chapter_id" TEXT NOT NULL,
    "last_page" INTEGER NOT NULL DEFAULT 0,
    "page_count" INTEGER NOT NULL DEFAULT 0,
    "completed" BOOLEAN NOT NULL DEFAULT false,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "reading_progress_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "extension" (
    "id" TEXT NOT NULL,
    "pkg_name" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "lang" TEXT NOT NULL,
    "icon_url" TEXT,
    "is_nsfw" BOOLEAN NOT NULL DEFAULT false,
    "installed_by_user_id" TEXT,
    "installed_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "health" "ExtensionHealth" NOT NULL DEFAULT 'OK',
    "consecutive_failures" INTEGER NOT NULL DEFAULT 0,
    "last_error_at" TIMESTAMP(3),
    "last_error_message" TEXT,

    CONSTRAINT "extension_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "user_extension_activation" (
    "user_id" TEXT NOT NULL,
    "extension_id" TEXT NOT NULL,
    "activated_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "user_extension_activation_pkey" PRIMARY KEY ("user_id","extension_id")
);

-- CreateTable
CREATE TABLE "extension_error_log" (
    "id" TEXT NOT NULL,
    "extension_id" TEXT NOT NULL,
    "occurred_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "message" TEXT NOT NULL,
    "context" TEXT,

    CONSTRAINT "extension_error_log_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "series" (
    "id" TEXT NOT NULL,
    "manga_id" INTEGER NOT NULL,
    "source_id" TEXT NOT NULL,
    "manga_url" TEXT NOT NULL,
    "title" TEXT NOT NULL,
    "thumbnail_url" TEXT,
    "type" "SeriesType" NOT NULL,
    "type_set_by_user_id" TEXT,
    "status" "SeriesStatus" NOT NULL DEFAULT 'UNKNOWN',
    "last_chapter_at" TIMESTAMP(3),
    "last_chapter_number" DOUBLE PRECISION,
    "resumed_at" TIMESTAMP(3),
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "series_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "chapter" (
    "id" TEXT NOT NULL,
    "chapter_id" INTEGER NOT NULL,
    "series_id" TEXT NOT NULL,
    "number" DOUBLE PRECISION NOT NULL,
    "name" TEXT NOT NULL,
    "scanlator" TEXT,
    "uploaded_at" TIMESTAMP(3) NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "chapter_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "user_reading_preference" (
    "user_id" TEXT NOT NULL,
    "type" "SeriesType" NOT NULL,
    "mode" "ReadingMode" NOT NULL,
    "direction" "ReadingDirection" NOT NULL,

    CONSTRAINT "user_reading_preference_pkey" PRIMARY KEY ("user_id","type")
);

-- CreateTable
CREATE TABLE "download_job" (
    "id" TEXT NOT NULL,
    "chapter_id" TEXT NOT NULL,
    "series_id" TEXT NOT NULL,
    "status" "DownloadStatus" NOT NULL DEFAULT 'PENDING',
    "priority" "DownloadPriority" NOT NULL DEFAULT 'BACKFILL',
    "attempts" INTEGER NOT NULL DEFAULT 0,
    "last_error" TEXT,
    "requested_by_user_id" TEXT,
    "next_attempt_at" TIMESTAMP(3),
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "download_job_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "app_user_email_key" ON "app_user"("email");

-- CreateIndex
CREATE INDEX "auth_identity_user_id_idx" ON "auth_identity"("user_id");

-- CreateIndex
CREATE UNIQUE INDEX "auth_identity_provider_subject_key" ON "auth_identity"("provider", "subject");

-- CreateIndex
CREATE INDEX "session_user_id_idx" ON "session"("user_id");

-- CreateIndex
CREATE UNIQUE INDEX "invite_code_key" ON "invite"("code");

-- CreateIndex
CREATE INDEX "subscription_user_id_idx" ON "subscription"("user_id");

-- CreateIndex
CREATE INDEX "subscription_series_id_idx" ON "subscription"("series_id");

-- CreateIndex
CREATE UNIQUE INDEX "subscription_user_id_series_id_key" ON "subscription"("user_id", "series_id");

-- CreateIndex
CREATE INDEX "category_user_id_idx" ON "category"("user_id");

-- CreateIndex
CREATE UNIQUE INDEX "category_user_id_name_key" ON "category"("user_id", "name");

-- CreateIndex
CREATE INDEX "subscription_category_category_id_idx" ON "subscription_category"("category_id");

-- CreateIndex
CREATE INDEX "reading_progress_user_id_idx" ON "reading_progress"("user_id");

-- CreateIndex
CREATE UNIQUE INDEX "reading_progress_user_id_chapter_id_key" ON "reading_progress"("user_id", "chapter_id");

-- CreateIndex
CREATE UNIQUE INDEX "extension_pkg_name_key" ON "extension"("pkg_name");

-- CreateIndex
CREATE INDEX "extension_installed_by_user_id_idx" ON "extension"("installed_by_user_id");

-- CreateIndex
CREATE INDEX "user_extension_activation_extension_id_idx" ON "user_extension_activation"("extension_id");

-- CreateIndex
CREATE INDEX "extension_error_log_extension_id_occurred_at_idx" ON "extension_error_log"("extension_id", "occurred_at");

-- CreateIndex
CREATE INDEX "series_last_chapter_at_idx" ON "series"("last_chapter_at");

-- CreateIndex
CREATE INDEX "series_type_idx" ON "series"("type");

-- CreateIndex
CREATE INDEX "series_status_idx" ON "series"("status");

-- CreateIndex
CREATE UNIQUE INDEX "series_manga_id_source_id_key" ON "series"("manga_id", "source_id");

-- CreateIndex
CREATE INDEX "chapter_series_id_uploaded_at_idx" ON "chapter"("series_id", "uploaded_at");

-- CreateIndex
CREATE UNIQUE INDEX "chapter_series_id_chapter_id_key" ON "chapter"("series_id", "chapter_id");

-- CreateIndex
CREATE INDEX "download_job_status_priority_created_at_idx" ON "download_job"("status", "priority", "created_at");

-- CreateIndex
CREATE INDEX "download_job_series_id_idx" ON "download_job"("series_id");

-- CreateIndex
CREATE UNIQUE INDEX "download_job_chapter_id_key" ON "download_job"("chapter_id");

-- AddForeignKey
ALTER TABLE "auth_identity" ADD CONSTRAINT "auth_identity_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "app_user"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "session" ADD CONSTRAINT "session_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "app_user"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "subscription" ADD CONSTRAINT "subscription_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "app_user"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "subscription" ADD CONSTRAINT "subscription_series_id_fkey" FOREIGN KEY ("series_id") REFERENCES "series"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "category" ADD CONSTRAINT "category_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "app_user"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "subscription_category" ADD CONSTRAINT "subscription_category_subscription_id_fkey" FOREIGN KEY ("subscription_id") REFERENCES "subscription"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "subscription_category" ADD CONSTRAINT "subscription_category_category_id_fkey" FOREIGN KEY ("category_id") REFERENCES "category"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "reading_progress" ADD CONSTRAINT "reading_progress_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "app_user"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "reading_progress" ADD CONSTRAINT "reading_progress_chapter_id_fkey" FOREIGN KEY ("chapter_id") REFERENCES "chapter"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "extension" ADD CONSTRAINT "extension_installed_by_user_id_fkey" FOREIGN KEY ("installed_by_user_id") REFERENCES "app_user"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "user_extension_activation" ADD CONSTRAINT "user_extension_activation_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "app_user"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "user_extension_activation" ADD CONSTRAINT "user_extension_activation_extension_id_fkey" FOREIGN KEY ("extension_id") REFERENCES "extension"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "extension_error_log" ADD CONSTRAINT "extension_error_log_extension_id_fkey" FOREIGN KEY ("extension_id") REFERENCES "extension"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "chapter" ADD CONSTRAINT "chapter_series_id_fkey" FOREIGN KEY ("series_id") REFERENCES "series"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "user_reading_preference" ADD CONSTRAINT "user_reading_preference_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "app_user"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "download_job" ADD CONSTRAINT "download_job_chapter_id_fkey" FOREIGN KEY ("chapter_id") REFERENCES "chapter"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "download_job" ADD CONSTRAINT "download_job_series_id_fkey" FOREIGN KEY ("series_id") REFERENCES "series"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "download_job" ADD CONSTRAINT "download_job_requested_by_user_id_fkey" FOREIGN KEY ("requested_by_user_id") REFERENCES "app_user"("id") ON DELETE SET NULL ON UPDATE CASCADE;
