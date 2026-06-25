-- CreateTable
CREATE TABLE "source" (
    "id" TEXT NOT NULL,
    "pkg_name" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "lang" TEXT NOT NULL,
    "is_nsfw" BOOLEAN NOT NULL DEFAULT false,
    "is_configurable" BOOLEAN NOT NULL DEFAULT false,
    "is_enabled" BOOLEAN NOT NULL DEFAULT true,

    CONSTRAINT "source_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE INDEX "source_pkg_name_idx" ON "source"("pkg_name");

-- AddForeignKey
ALTER TABLE "source" ADD CONSTRAINT "source_pkg_name_fkey" FOREIGN KEY ("pkg_name") REFERENCES "extension"("pkg_name") ON DELETE CASCADE ON UPDATE CASCADE;
