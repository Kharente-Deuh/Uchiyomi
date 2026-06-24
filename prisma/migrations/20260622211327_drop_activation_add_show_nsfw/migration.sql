/*
  Warnings:

  - You are about to drop the `user_extension_activation` table. If the table is not empty, all the data it contains will be lost.

*/
-- DropForeignKey
ALTER TABLE "user_extension_activation" DROP CONSTRAINT "user_extension_activation_extension_id_fkey";

-- DropForeignKey
ALTER TABLE "user_extension_activation" DROP CONSTRAINT "user_extension_activation_user_id_fkey";

-- AlterTable
ALTER TABLE "app_user" ADD COLUMN     "show_nsfw" BOOLEAN NOT NULL DEFAULT false;

-- DropTable
DROP TABLE "user_extension_activation";
