mysqldump: [Warning] Using a password on the command line interface can be insecure.
-- MySQL dump 10.13  Distrib 8.0.45, for Linux (x86_64)
--
-- Host: localhost    Database: bridge_detection
-- ------------------------------------------------------
-- Server version	8.0.45-0ubuntu0.22.04.1

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `bridges`
--

DROP TABLE IF EXISTS `bridges`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `bridges` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `bridge_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `bridge_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `address` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `longitude` decimal(10,6) DEFAULT NULL,
  `latitude` decimal(10,6) DEFAULT NULL,
  `bridge_type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `build_year` bigint DEFAULT NULL,
  `length` decimal(10,2) DEFAULT NULL,
  `width` decimal(10,2) DEFAULT NULL,
  `status` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT '正常',
  `remark` text COLLATE utf8mb4_unicode_ci,
  `deleted_at` datetime(3) DEFAULT NULL,
  `model_3d_path` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_bridges_bridge_code` (`bridge_code`),
  KEY `idx_bridges_user_id` (`user_id`),
  KEY `idx_bridges_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_bridges_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=27 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `bridges`
--

LOCK TABLES `bridges` WRITE;
/*!40000 ALTER TABLE `bridges` DISABLE KEYS */;
INSERT INTO `bridges` VALUES (1,3,'2026-03-16 16:00:41.598','2026-03-16 16:00:41.852','黄河大桥（已更新）','HH001_deleted_1773648041','河南省郑州市',113.625000,34.746000,'悬索桥',1985,5464.00,23.00,'正常','测试更新','2026-03-16 16:00:41.853',''),(2,7,'2026-03-17 17:47:50.549','2026-03-17 17:47:50.549','长江大桥_1773740870','BRIDGE_CJ_1773740870','湖北省武汉市',114.310000,30.520000,'悬索桥',1957,1670.40,18.00,'正常','武汉长江大桥',NULL,''),(3,7,'2026-03-17 17:47:50.606','2026-03-17 17:47:50.606','黄河大桥_1773740870','BRIDGE_HH_1773740870','河南省郑州市',113.620000,34.750000,'斜拉桥',1986,2386.00,22.50,'正常','郑州黄河大桥',NULL,''),(4,7,'2026-03-17 17:47:50.664','2026-03-17 17:47:50.664','珠江大桥_1773740870','BRIDGE_ZJ_1773740870','广东省广州市',113.260000,23.130000,'梁桥',2008,1388.00,28.00,'正常','广州珠江大桥',NULL,''),(5,7,'2026-03-17 17:48:03.210','2026-03-17 17:48:03.210','测试桥梁','TEST_BRIDGE_999','测试地址',118.780000,32.040000,'梁桥',2020,1000.00,20.00,'正常','测试',NULL,''),(6,8,'2026-03-17 17:48:58.799','2026-03-17 17:48:58.799','长江大桥_1773740938','BRIDGE_CJ_1773740938','湖北省武汉市',114.310000,30.520000,'悬索桥',1957,1670.40,18.00,'正常','武汉长江大桥',NULL,''),(7,8,'2026-03-17 17:48:58.855','2026-03-17 17:48:58.855','黄河大桥_1773740938','BRIDGE_HH_1773740938','河南省郑州市',113.620000,34.750000,'斜拉桥',1986,2386.00,22.50,'正常','郑州黄河大桥',NULL,''),(8,8,'2026-03-17 17:48:58.912','2026-03-17 17:48:59.687','珠江大桥_1773740938','BRIDGE_ZJ_1773740938','广东省广州市',113.260000,23.130000,'梁桥',2008,1388.00,28.00,'维修中','发现轻微缺陷，正在进行维护',NULL,''),(9,9,'2026-03-17 17:49:08.590','2026-03-17 17:49:08.590','长江大桥_1773740948','BRIDGE_CJ_1773740948','湖北省武汉市',114.310000,30.520000,'悬索桥',1957,1670.40,18.00,'正常','武汉长江大桥',NULL,''),(10,9,'2026-03-17 17:49:08.664','2026-03-17 17:49:08.664','黄河大桥_1773740948','BRIDGE_HH_1773740948','河南省郑州市',113.620000,34.750000,'斜拉桥',1986,2386.00,22.50,'正常','郑州黄河大桥',NULL,''),(11,9,'2026-03-17 17:49:08.730','2026-03-17 17:49:09.508','珠江大桥_1773740948','BRIDGE_ZJ_1773740948','广东省广州市',113.260000,23.130000,'梁桥',2008,1388.00,28.00,'维修中','发现轻微缺陷，正在进行维护',NULL,''),(12,10,'2026-03-17 17:50:14.479','2026-03-17 17:50:14.479','长江大桥_1773741014','BRIDGE_CJ_1773741014','湖北省武汉市',114.310000,30.520000,'悬索桥',1957,1670.40,18.00,'正常','武汉长江大桥',NULL,''),(13,10,'2026-03-17 17:50:14.541','2026-03-17 17:50:14.541','黄河大桥_1773741014','BRIDGE_HH_1773741014','河南省郑州市',113.620000,34.750000,'斜拉桥',1986,2386.00,22.50,'正常','郑州黄河大桥',NULL,''),(14,10,'2026-03-17 17:50:14.594','2026-03-17 17:50:15.365','珠江大桥_1773741014','BRIDGE_ZJ_1773741014','广东省广州市',113.260000,23.130000,'梁桥',2008,1388.00,28.00,'维修中','发现轻微缺陷，正在进行维护',NULL,''),(15,13,'2026-03-17 23:32:07.421','2026-03-17 23:32:07.421','长江大桥','BRIDGE_CJ_1773761527','湖北省武汉市',114.310000,30.590000,'悬索桥',1957,1670.00,18.00,'正常','测试桥梁',NULL,''),(16,14,'2026-03-18 00:02:03.825','2026-03-18 00:02:03.825','测试桥梁','BRIDGE_TEST_1773763323','测试地址',114.310000,30.590000,'梁桥',2020,100.00,20.00,'正常','测试',NULL,''),(17,15,'2026-03-18 00:16:30.892','2026-03-18 00:16:32.050','长江大桥','BRIDGE_CJ_1773764190_deleted_1773764190','湖北省武汉市',114.310000,30.590000,'悬索桥',1957,1670.00,18.00,'正常','测试桥梁','2026-03-18 00:16:32.050',''),(18,16,'2026-03-18 00:18:14.054','2026-03-18 00:18:15.202','长江大桥','BRIDGE_CJ_1773764293_deleted_1773764294','湖北省武汉市',114.310000,30.590000,'悬索桥',1957,1670.00,18.00,'正常','测试桥梁','2026-03-18 00:18:15.202',''),(19,17,'2026-03-18 00:20:57.440','2026-03-18 00:20:58.608','长江大桥','BRIDGE_CJ_1773764457_deleted_1773764457','湖北省武汉市',114.310000,30.590000,'悬索桥',1957,1670.00,18.00,'正常','测试桥梁','2026-03-18 00:20:58.609',''),(20,18,'2026-03-18 00:22:33.397','2026-03-18 00:22:34.545','长江大桥','BRIDGE_CJ_1773764553_deleted_1773764553','湖北省武汉市',114.310000,30.590000,'悬索桥',1957,1670.00,18.00,'正常','测试桥梁','2026-03-18 00:22:34.546',''),(21,19,'2026-03-18 00:23:19.939','2026-03-18 00:23:21.093','长江大桥','BRIDGE_CJ_1773764599_deleted_1773764599','湖北省武汉市',114.310000,30.590000,'悬索桥',1957,1670.00,18.00,'正常','测试桥梁','2026-03-18 00:23:21.093',''),(22,20,'2026-03-18 00:23:54.557','2026-03-18 00:23:55.715','长江大桥','BRIDGE_CJ_1773764634_deleted_1773764634','湖北省武汉市',114.310000,30.590000,'悬索桥',1957,1670.00,18.00,'正常','测试桥梁','2026-03-18 00:23:55.716',''),(23,21,'2026-03-18 01:15:11.345','2026-03-18 01:15:12.591','长江大桥','BRIDGE_CJ_1773767711_deleted_1773767711','湖北省武汉市',114.310000,30.590000,'悬索桥',1957,1670.00,18.00,'正常','测试桥梁','2026-03-18 01:15:12.592',''),(24,22,'2026-03-18 01:18:22.541','2026-03-18 01:18:23.777','长江大桥','BRIDGE_CJ_1773767902_deleted_1773767902','湖北省武汉市',114.310000,30.590000,'悬索桥',1957,1670.00,18.00,'正常','测试桥梁','2026-03-18 01:18:23.778',''),(25,24,'2026-03-18 01:19:55.913','2026-03-18 01:19:57.169','长江大桥','BRIDGE_CJ_1773767995_deleted_1773767995','湖北省武汉市',114.310000,30.590000,'悬索桥',1957,1670.00,18.00,'正常','测试桥梁','2026-03-18 01:19:57.170',''),(26,25,'2026-03-18 01:20:04.651','2026-03-18 01:20:05.901','长江大桥','BRIDGE_CJ_1773768004_deleted_1773768004','湖北省武汉市',114.310000,30.590000,'悬索桥',1957,1670.00,18.00,'正常','测试桥梁','2026-03-18 01:20:05.902','');
/*!40000 ALTER TABLE `bridges` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `defects`
--

DROP TABLE IF EXISTS `defects`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `defects` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `bridge_id` bigint unsigned NOT NULL,
  `defect_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `image_path` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `result_path` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `b_box` text COLLATE utf8mb4_unicode_ci,
  `length` decimal(10,4) DEFAULT NULL,
  `width` decimal(10,4) DEFAULT NULL,
  `area` decimal(10,4) DEFAULT NULL,
  `confidence` decimal(5,4) DEFAULT NULL,
  `detected_at` datetime NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_defects_bridge_id` (`bridge_id`),
  KEY `idx_defects_detected_at` (`detected_at`),
  KEY `idx_defects_deleted_at` (`deleted_at`),
  KEY `idx_defects_type` (`defect_type`),
  KEY `idx_defects_bridge_time` (`bridge_id`,`detected_at`),
  KEY `idx_defects_confidence` (`confidence`),
  CONSTRAINT `fk_bridges_defects` FOREIGN KEY (`bridge_id`) REFERENCES `bridges` (`id`),
  CONSTRAINT `fk_defects_bridge` FOREIGN KEY (`bridge_id`) REFERENCES `bridges` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=52 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `defects`
--

LOCK TABLES `defects` WRITE;
/*!40000 ALTER TABLE `defects` DISABLE KEYS */;
INSERT INTO `defects` VALUES (1,6,'裂缝','images/cj_crack_001.jpg','results/cj_crack_001_result.jpg',NULL,2.5000,0.0600,0.1500,0.9600,'2026-03-16 17:48:59','2026-03-17 17:48:59.000',NULL),(2,6,'裂缝','images/cj_crack_002.jpg','results/cj_crack_002_result.jpg',NULL,1.8000,0.0440,0.0800,0.9400,'2026-03-15 17:48:59','2026-03-17 17:48:59.000',NULL),(3,6,'剥落','images/cj_spall_001.jpg','results/cj_spall_001_result.jpg',NULL,1.5000,0.0800,0.1200,0.9200,'2026-03-16 17:48:59','2026-03-17 17:48:59.000',NULL),(4,6,'钢筋锈蚀','images/cj_rust_001.jpg','results/cj_rust_001_result.jpg',NULL,0.8000,0.0630,0.0500,0.8900,'2026-03-14 17:48:59','2026-03-17 17:48:59.000',NULL),(5,6,'混凝土开裂','images/cj_concrete_001.jpg','results/cj_concrete_001_result.jpg',NULL,1.2000,0.0500,0.0600,0.8700,'2026-03-15 17:48:59','2026-03-17 17:48:59.000',NULL),(6,6,'裂缝','images/cj_crack_003.jpg','results/cj_crack_003_result.jpg',NULL,0.9000,0.0440,0.0400,0.8500,'2026-03-13 17:48:59','2026-03-17 17:48:59.000',NULL),(7,6,'表面损伤','images/cj_damage_001.jpg','results/cj_damage_001_result.jpg',NULL,0.5000,0.0400,0.0200,0.7800,'2026-03-12 17:48:59','2026-03-17 17:48:59.000',NULL),(8,6,'裂缝','images/cj_crack_004.jpg','results/cj_crack_004_result.jpg',NULL,0.7000,0.0430,0.0300,0.8800,'2026-03-17 17:48:59','2026-03-17 17:48:59.000',NULL),(9,6,'剥落','images/cj_spall_002.jpg','results/cj_spall_002_result.jpg',NULL,1.1000,0.0640,0.0700,0.9100,'2026-03-17 17:48:59','2026-03-17 17:48:59.000',NULL),(10,6,'钢筋锈蚀','images/cj_rust_002.jpg','results/cj_rust_002_result.jpg',NULL,0.6000,0.0670,0.0400,0.8600,'2026-03-16 17:48:59','2026-03-17 17:48:59.000',NULL),(11,7,'裂缝','images/hh_crack_001.jpg','results/hh_crack_001_result.jpg',NULL,1.2000,0.0500,0.0600,0.9200,'2026-03-16 17:48:59','2026-03-17 17:48:59.000',NULL),(12,7,'裂缝','images/hh_crack_002.jpg','results/hh_crack_002_result.jpg',NULL,0.9000,0.0440,0.0400,0.8800,'2026-03-15 17:48:59','2026-03-17 17:48:59.000',NULL),(13,7,'剥落','images/hh_spall_001.jpg','results/hh_spall_001_result.jpg',NULL,0.7000,0.0430,0.0300,0.8500,'2026-03-14 17:48:59','2026-03-17 17:48:59.000',NULL),(14,7,'混凝土开裂','images/hh_concrete_001.jpg','results/hh_concrete_001_result.jpg',NULL,0.5000,0.0400,0.0200,0.8200,'2026-03-13 17:48:59','2026-03-17 17:48:59.000',NULL),(15,7,'表面损伤','images/hh_damage_001.jpg','results/hh_damage_001_result.jpg',NULL,0.4000,0.0380,0.0150,0.7500,'2026-03-17 17:48:59','2026-03-17 17:48:59.000',NULL),(16,8,'裂缝','images/zj_crack_001.jpg','results/zj_crack_001_result.jpg',NULL,0.5000,0.0400,0.0200,0.7800,'2026-03-16 17:48:59','2026-03-17 17:48:59.000',NULL),(17,8,'表面损伤','images/zj_damage_001.jpg','results/zj_damage_001_result.jpg',NULL,0.3000,0.0330,0.0100,0.7200,'2026-03-17 17:48:59','2026-03-17 17:48:59.000',NULL),(18,9,'裂缝','images/cj_crack_001.jpg','results/cj_crack_001_result.jpg',NULL,2.5000,0.0600,0.1500,0.9600,'2026-03-16 17:49:08','2026-03-17 17:49:08.000',NULL),(19,9,'裂缝','images/cj_crack_002.jpg','results/cj_crack_002_result.jpg',NULL,1.8000,0.0440,0.0800,0.9400,'2026-03-15 17:49:08','2026-03-17 17:49:08.000',NULL),(20,9,'剥落','images/cj_spall_001.jpg','results/cj_spall_001_result.jpg',NULL,1.5000,0.0800,0.1200,0.9200,'2026-03-16 17:49:08','2026-03-17 17:49:08.000',NULL),(21,9,'钢筋锈蚀','images/cj_rust_001.jpg','results/cj_rust_001_result.jpg',NULL,0.8000,0.0630,0.0500,0.8900,'2026-03-14 17:49:08','2026-03-17 17:49:08.000',NULL),(22,9,'混凝土开裂','images/cj_concrete_001.jpg','results/cj_concrete_001_result.jpg',NULL,1.2000,0.0500,0.0600,0.8700,'2026-03-15 17:49:08','2026-03-17 17:49:08.000',NULL),(23,9,'裂缝','images/cj_crack_003.jpg','results/cj_crack_003_result.jpg',NULL,0.9000,0.0440,0.0400,0.8500,'2026-03-13 17:49:08','2026-03-17 17:49:08.000',NULL),(24,9,'表面损伤','images/cj_damage_001.jpg','results/cj_damage_001_result.jpg',NULL,0.5000,0.0400,0.0200,0.7800,'2026-03-12 17:49:08','2026-03-17 17:49:08.000',NULL),(25,9,'裂缝','images/cj_crack_004.jpg','results/cj_crack_004_result.jpg',NULL,0.7000,0.0430,0.0300,0.8800,'2026-03-17 17:49:08','2026-03-17 17:49:08.000',NULL),(26,9,'剥落','images/cj_spall_002.jpg','results/cj_spall_002_result.jpg',NULL,1.1000,0.0640,0.0700,0.9100,'2026-03-17 17:49:08','2026-03-17 17:49:08.000',NULL),(27,9,'钢筋锈蚀','images/cj_rust_002.jpg','results/cj_rust_002_result.jpg',NULL,0.6000,0.0670,0.0400,0.8600,'2026-03-16 17:49:08','2026-03-17 17:49:08.000',NULL),(28,10,'裂缝','images/hh_crack_001.jpg','results/hh_crack_001_result.jpg',NULL,1.2000,0.0500,0.0600,0.9200,'2026-03-16 17:49:08','2026-03-17 17:49:08.000',NULL),(29,10,'裂缝','images/hh_crack_002.jpg','results/hh_crack_002_result.jpg',NULL,0.9000,0.0440,0.0400,0.8800,'2026-03-15 17:49:08','2026-03-17 17:49:08.000',NULL),(30,10,'剥落','images/hh_spall_001.jpg','results/hh_spall_001_result.jpg',NULL,0.7000,0.0430,0.0300,0.8500,'2026-03-14 17:49:08','2026-03-17 17:49:08.000',NULL),(31,10,'混凝土开裂','images/hh_concrete_001.jpg','results/hh_concrete_001_result.jpg',NULL,0.5000,0.0400,0.0200,0.8200,'2026-03-13 17:49:08','2026-03-17 17:49:08.000',NULL),(32,10,'表面损伤','images/hh_damage_001.jpg','results/hh_damage_001_result.jpg',NULL,0.4000,0.0380,0.0150,0.7500,'2026-03-17 17:49:08','2026-03-17 17:49:08.000',NULL),(33,11,'裂缝','images/zj_crack_001.jpg','results/zj_crack_001_result.jpg',NULL,0.5000,0.0400,0.0200,0.7800,'2026-03-16 17:49:08','2026-03-17 17:49:08.000',NULL),(34,11,'表面损伤','images/zj_damage_001.jpg','results/zj_damage_001_result.jpg',NULL,0.3000,0.0330,0.0100,0.7200,'2026-03-17 17:49:08','2026-03-17 17:49:08.000',NULL),(35,12,'裂缝','images/cj_crack_001.jpg','results/cj_crack_001_result.jpg',NULL,2.5000,0.0600,0.1500,0.9600,'2026-03-16 17:50:14','2026-03-17 17:50:14.000',NULL),(36,12,'裂缝','images/cj_crack_002.jpg','results/cj_crack_002_result.jpg',NULL,1.8000,0.0440,0.0800,0.9400,'2026-03-15 17:50:14','2026-03-17 17:50:14.000',NULL),(37,12,'剥落','images/cj_spall_001.jpg','results/cj_spall_001_result.jpg',NULL,1.5000,0.0800,0.1200,0.9200,'2026-03-16 17:50:14','2026-03-17 17:50:14.000',NULL),(38,12,'钢筋锈蚀','images/cj_rust_001.jpg','results/cj_rust_001_result.jpg',NULL,0.8000,0.0630,0.0500,0.8900,'2026-03-14 17:50:14','2026-03-17 17:50:14.000',NULL),(39,12,'混凝土开裂','images/cj_concrete_001.jpg','results/cj_concrete_001_result.jpg',NULL,1.2000,0.0500,0.0600,0.8700,'2026-03-15 17:50:14','2026-03-17 17:50:14.000',NULL),(40,12,'裂缝','images/cj_crack_003.jpg','results/cj_crack_003_result.jpg',NULL,0.9000,0.0440,0.0400,0.8500,'2026-03-13 17:50:14','2026-03-17 17:50:14.000',NULL),(41,12,'表面损伤','images/cj_damage_001.jpg','results/cj_damage_001_result.jpg',NULL,0.5000,0.0400,0.0200,0.7800,'2026-03-12 17:50:14','2026-03-17 17:50:14.000',NULL),(42,12,'裂缝','images/cj_crack_004.jpg','results/cj_crack_004_result.jpg',NULL,0.7000,0.0430,0.0300,0.8800,'2026-03-17 17:50:14','2026-03-17 17:50:14.000',NULL),(43,12,'剥落','images/cj_spall_002.jpg','results/cj_spall_002_result.jpg',NULL,1.1000,0.0640,0.0700,0.9100,'2026-03-17 17:50:14','2026-03-17 17:50:14.000',NULL),(44,12,'钢筋锈蚀','images/cj_rust_002.jpg','results/cj_rust_002_result.jpg',NULL,0.6000,0.0670,0.0400,0.8600,'2026-03-16 17:50:14','2026-03-17 17:50:14.000',NULL),(45,13,'裂缝','images/hh_crack_001.jpg','results/hh_crack_001_result.jpg',NULL,1.2000,0.0500,0.0600,0.9200,'2026-03-16 17:50:14','2026-03-17 17:50:14.000',NULL),(46,13,'裂缝','images/hh_crack_002.jpg','results/hh_crack_002_result.jpg',NULL,0.9000,0.0440,0.0400,0.8800,'2026-03-15 17:50:14','2026-03-17 17:50:14.000',NULL),(47,13,'剥落','images/hh_spall_001.jpg','results/hh_spall_001_result.jpg',NULL,0.7000,0.0430,0.0300,0.8500,'2026-03-14 17:50:14','2026-03-17 17:50:14.000',NULL),(48,13,'混凝土开裂','images/hh_concrete_001.jpg','results/hh_concrete_001_result.jpg',NULL,0.5000,0.0400,0.0200,0.8200,'2026-03-13 17:50:14','2026-03-17 17:50:14.000',NULL),(49,13,'表面损伤','images/hh_damage_001.jpg','results/hh_damage_001_result.jpg',NULL,0.4000,0.0380,0.0150,0.7500,'2026-03-17 17:50:14','2026-03-17 17:50:14.000',NULL),(50,14,'裂缝','images/zj_crack_001.jpg','results/zj_crack_001_result.jpg',NULL,0.5000,0.0400,0.0200,0.7800,'2026-03-16 17:50:14','2026-03-17 17:50:14.000',NULL),(51,14,'表面损伤','images/zj_damage_001.jpg','results/zj_damage_001_result.jpg',NULL,0.3000,0.0330,0.0100,0.7200,'2026-03-17 17:50:14','2026-03-17 17:50:14.000',NULL);
/*!40000 ALTER TABLE `defects` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `drones`
--

DROP TABLE IF EXISTS `drones`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `drones` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `model` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `stream_url` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `user_id` bigint unsigned NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_drones_user_id` (`user_id`),
  CONSTRAINT `fk_drones_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `drones`
--

LOCK TABLES `drones` WRITE;
/*!40000 ALTER TABLE `drones` DISABLE KEYS */;
INSERT INTO `drones` VALUES (3,'大疆 Phantom 4 RTK_1773740814','Phantom 4 RTK','rtsp://192.168.1.101:554/stream2',6,'2026-03-17 17:46:55.059','2026-03-17 17:46:55.059'),(5,'大疆 Phantom 4 RTK_1773740870','Phantom 4 RTK','rtsp://192.168.1.101:554/stream2',7,'2026-03-17 17:47:50.824','2026-03-17 17:47:50.824'),(7,'大疆 Phantom 4 RTK_1773740938','Phantom 4 RTK','rtsp://192.168.1.101:554/stream2',8,'2026-03-17 17:48:59.072','2026-03-17 17:48:59.072'),(9,'大疆 Phantom 4 RTK_1773740948','Phantom 4 RTK','rtsp://192.168.1.101:554/stream2',9,'2026-03-17 17:49:08.890','2026-03-17 17:49:08.890'),(11,'大疆 Phantom 4 RTK_1773741014','Phantom 4 RTK','rtsp://192.168.1.101:554/stream2',10,'2026-03-17 17:50:14.755','2026-03-17 17:50:14.755');
/*!40000 ALTER TABLE `drones` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `reports`
--

DROP TABLE IF EXISTS `reports`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `reports` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '报表ID（主键）',
  `report_name` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '报表名称',
  `report_type` enum('bridge_inspection','defect_analysis','health_comparison') COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '报表类型',
  `user_id` bigint unsigned NOT NULL COMMENT '创建用户ID（外键）',
  `bridge_id` bigint unsigned DEFAULT NULL COMMENT '关联桥梁ID（单桥梁报表）',
  `bridge_ids` json DEFAULT NULL COMMENT '关联桥梁ID列表（多桥梁报表）',
  `start_time` datetime NOT NULL COMMENT '报表开始时间',
  `end_time` datetime NOT NULL COMMENT '报表结束时间',
  `file_path` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'PDF文件路径',
  `file_size` bigint DEFAULT '0' COMMENT '文件大小（字节）',
  `status` enum('generating','completed','failed') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'generating' COMMENT '生成状态',
  `error_message` text COLLATE utf8mb4_unicode_ci COMMENT '错误信息（失败时）',
  `total_pages` int DEFAULT '0' COMMENT '总页数',
  `defect_count` int DEFAULT '0' COMMENT '缺陷数量',
  `high_risk_count` int DEFAULT '0' COMMENT '高危缺陷数量',
  `health_score` decimal(5,2) DEFAULT '0.00' COMMENT '健康度评分',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间（软删除）',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`) COMMENT '用户ID索引',
  KEY `idx_bridge_id` (`bridge_id`) COMMENT '桥梁ID索引',
  KEY `idx_report_type` (`report_type`) COMMENT '报表类型索引',
  KEY `idx_status` (`status`) COMMENT '状态索引',
  KEY `idx_created_at` (`created_at`) COMMENT '创建时间索引',
  KEY `idx_deleted_at` (`deleted_at`) COMMENT '软删除索引',
  CONSTRAINT `fk_reports_bridge_id` FOREIGN KEY (`bridge_id`) REFERENCES `bridges` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `fk_reports_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='检测报表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `reports`
--

LOCK TABLES `reports` WRITE;
/*!40000 ALTER TABLE `reports` DISABLE KEYS */;
INSERT INTO `reports` VALUES (1,'长江大桥检测报表_1773764190','bridge_inspection',15,17,NULL,'2026-02-16 08:00:00','2026-03-18 08:00:00','',0,'failed','PDF生成失败：保存PDF文件失败: not supported\n ',0,0,0,100.00,'2026-03-17 16:16:31','2026-03-17 16:16:32','2026-03-17 16:16:32'),(2,'长江大桥检测报表_1773764293','bridge_inspection',16,18,NULL,'2026-02-16 08:00:00','2026-03-18 08:00:00','',0,'failed','PDF生成失败：保存PDF文件失败: not supported\n ',0,0,0,100.00,'2026-03-17 16:18:14','2026-03-17 16:18:15','2026-03-17 16:18:15'),(3,'长江大桥检测报表_1773764457','bridge_inspection',17,19,NULL,'2026-02-16 08:00:00','2026-03-18 08:00:00','',0,'failed','PDF生成失败：保存PDF文件失败: not supported\n ',0,0,0,100.00,'2026-03-17 16:20:57','2026-03-17 16:20:58','2026-03-17 16:20:59'),(4,'长江大桥检测报表_1773764553','bridge_inspection',18,20,NULL,'2026-02-16 08:00:00','2026-03-18 08:00:00','',0,'failed','PDF生成失败：保存PDF文件失败: not supported\n ',0,0,0,100.00,'2026-03-17 16:22:33','2026-03-17 16:22:34','2026-03-17 16:22:35'),(5,'长江大桥检测报表_1773764599','bridge_inspection',19,21,NULL,'2026-02-16 08:00:00','2026-03-18 08:00:00','',0,'failed','PDF生成失败：保存PDF文件失败: not supported\n ',0,0,0,100.00,'2026-03-17 16:23:20','2026-03-17 16:23:21','2026-03-17 16:23:21'),(6,'长江大桥检测报表_1773764634','bridge_inspection',20,22,NULL,'2026-02-16 08:00:00','2026-03-18 08:00:00','',0,'failed','PDF生成失败：保存PDF文件失败: not supported\n ',0,0,0,100.00,'2026-03-17 16:23:55','2026-03-17 16:23:55','2026-03-17 16:23:56'),(7,'长江大桥检测报表_1773767711','bridge_inspection',21,23,NULL,'2026-02-16 08:00:00','2026-03-18 08:00:00','',0,'failed','PDF生成失败：添加中文字体失败: 加载字体文件失败: Unrecognized file (font) format',0,0,0,100.00,'2026-03-17 17:15:11','2026-03-17 17:15:12','2026-03-17 17:15:13'),(8,'长江大桥检测报表_1773767902','bridge_inspection',22,24,NULL,'2026-02-16 08:00:00','2026-03-18 08:00:00','',0,'failed','PDF生成失败：添加中文字体失败: 加载字体文件失败: Unrecognized file (font) format',0,0,0,100.00,'2026-03-17 17:18:23','2026-03-17 17:18:23','2026-03-17 17:18:24'),(9,'长江大桥检测报表_1773767995','bridge_inspection',24,25,NULL,'2026-02-16 08:00:00','2026-03-18 08:00:00','reports/report_9_20260318011955.pdf',17617,'completed','',4,0,0,100.00,'2026-03-17 17:19:56','2026-03-17 17:19:57','2026-03-17 17:19:57'),(10,'长江大桥检测报表_1773768004','bridge_inspection',25,26,NULL,'2026-02-16 08:00:00','2026-03-18 08:00:00','reports/report_10_20260318012004.pdf',17624,'completed','',4,0,0,100.00,'2026-03-17 17:20:05','2026-03-17 17:20:05','2026-03-17 17:20:06');
/*!40000 ALTER TABLE `reports` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `password` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `email` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `role` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'user',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `real_name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_users_username` (`username`),
  KEY `idx_users_email` (`email`)
) ENGINE=InnoDB AUTO_INCREMENT=26 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `users`
--

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
INSERT INTO `users` VALUES (2,'testuser','$2a$10$AkJ4CFwA/skpzztdcfZt5uLDw7AggYMITiIii0KcwyovxI4.nC4HG','test@example.com','user','2026-03-15 18:40:50.483','2026-03-15 18:42:41.023','更新后的名字','13900139000'),(3,'admin','$2a$10$uj7v93/2yLlkdSWmXTxPbe6GFM71.IlyRa.bGsVgjsVUCPWjGtncy','admin@example.com','admin','2026-03-15 18:42:27.607','2026-03-16 15:57:47.552','管理员（已更新）',''),(4,'testuser1','$2a$10$0JfVx5j7DAdi2fVV4b/2LOrvofGk5WxYLwRXGNfYlQmhZD7Q3Vp/i','testuser1@example.com','user','2026-03-16 17:29:50.418','2026-03-16 17:29:50.418','测试用户1',''),(5,'testuser2','$2a$10$I43nwSt7O/6ONvQvA.upJ.IT.nQbTmKcDm30N2wNqvCKs.pabt/wS','testuser2@example.com','user','2026-03-16 17:32:12.959','2026-03-16 17:32:12.959','测试用户2',''),(6,'user_test_1773740814','$2a$10$xB9ul.F7pGJ7lSYA48vO6umEjn4AO/mfHZsNk69TwR.VrLJJDwZce','test_1773740814@example.com','user','2026-03-17 17:46:54.640','2026-03-17 17:46:54.640','测试用户','13800138000'),(7,'user_test_1773740870','$2a$10$c.B9/yaycsf9DkQPzsA6He4AFWYJFWH6SA7hsMNey8jyRdkDpKemu','test_1773740870@example.com','user','2026-03-17 17:47:50.401','2026-03-17 17:47:50.401','测试用户','13800138000'),(8,'user_test_1773740938','$2a$10$jWI612AcYj04wn82X4OGaOyCYW08kA..oeNWyEPU//l0TSd1BSm5e','test_1773740938@example.com','user','2026-03-17 17:48:58.633','2026-03-17 17:48:58.633','测试用户','13800138000'),(9,'user_test_1773740948','$2a$10$e0CaG2g3aqXS5TKCQckR5.Rggl2TJZaGajYcjDLYiv.nTqLhgQk4u','test_1773740948@example.com','user','2026-03-17 17:49:08.431','2026-03-17 17:49:08.431','测试用户','13800138000'),(10,'user_test_1773741014','$2a$10$BQpQ14hK/dTTndecjwZviuvONxevzH8Hfy3GZad2FvscwapY1zQiG','test_1773741014@example.com','user','2026-03-17 17:50:14.330','2026-03-17 17:50:14.330','测试用户','13800138000'),(11,'report_test_1773761236','$2a$10$TfNXnKrDNpzozwkin/mLtuRGOR5MU9JYjCXPlO6UESsKnacIe4N5i','report_test_1773761236@test.com','user','2026-03-17 23:27:16.335','2026-03-17 23:27:16.335','报表测试用户',''),(12,'report_test_1773761484','$2a$10$YSmb6AF9t5XnXGEMkcDTCeFULwuK1s9Bu/XteNS.v7nEB78hxtIRq','report_test_1773761484@test.com','user','2026-03-17 23:31:24.250','2026-03-17 23:31:24.250','报表测试用户',''),(13,'report_test_1773761527','$2a$10$cwow77PVfB6GSDrNxM4qv.j9yU.TkE31WsYmR.ZBFw5awDfPAwMj6','report_test_1773761527@test.com','user','2026-03-17 23:32:07.336','2026-03-17 23:32:07.336','报表测试用户',''),(14,'debug_1773763323','$2a$10$d9n2EfEAJSIzU838v2yPM.my6qbu.ZsOzw0PKeTLP6h.GMH5Mzxfu','debug_1773763323@test.com','user','2026-03-18 00:02:03.733','2026-03-18 00:02:03.733','调试用户',''),(15,'report_test_1773764190','$2a$10$kfrolo68JNuYqQAUPA61O.kY79HfOwiaeWSSyenapvCHf89qyETD6','report_test_1773764190@test.com','user','2026-03-18 00:16:30.799','2026-03-18 00:16:30.799','报表测试用户',''),(16,'report_test_1773764293','$2a$10$2iyMVw9IEHN2h8zURD26xuAvO9EmfwP550lb.qYvXE0nfNNM0jpwG','report_test_1773764293@test.com','user','2026-03-18 00:18:13.966','2026-03-18 00:18:13.966','报表测试用户',''),(17,'report_test_1773764457','$2a$10$2hitJECISw9kOxJMEFNGbes9GhDi9kyrZmsZcIUuNPS3aJ8xFihSS','report_test_1773764457@test.com','user','2026-03-18 00:20:57.351','2026-03-18 00:20:57.351','报表测试用户',''),(18,'report_test_1773764553','$2a$10$HcMS8XQm.7LEEeyI0RoKxuqNwVJyPuESgDtrFryDz5dxjRxBXX7Uy','report_test_1773764553@test.com','user','2026-03-18 00:22:33.305','2026-03-18 00:22:33.305','报表测试用户',''),(19,'report_test_1773764599','$2a$10$GtwA9aaEtqKpMSj81MK7HO2.f3wfN9.sLvYndBZrUAU04MicMxwrW','report_test_1773764599@test.com','user','2026-03-18 00:23:19.854','2026-03-18 00:23:19.854','报表测试用户',''),(20,'report_test_1773764634','$2a$10$dKSXDldyQiTgXacxUImJcO.sFuDKoq46XmQ7851IO.mOUbb1RO3Vm','report_test_1773764634@test.com','user','2026-03-18 00:23:54.442','2026-03-18 00:23:54.442','报表测试用户',''),(21,'report_test_1773767711','$2a$10$HOcMlH8aSniVR7Jc3.8nveie3hjzXnP3UHIho/KpsPr2NnzOqqAje','report_test_1773767711@test.com','user','2026-03-18 01:15:11.172','2026-03-18 01:15:11.172','报表测试用户',''),(22,'report_test_1773767902','$2a$10$8KZ2LvkqC2kRvoebgINhoecDRzT.Qs6LnKOUjeJAgyx3x5nua5dlq','report_test_1773767902@test.com','user','2026-03-18 01:18:22.424','2026-03-18 01:18:22.424','报表测试用户',''),(23,'quicktest','$2a$10$WD9pZ6RNbkfONgB5XAtmvOjTCGJiQj8Z0IvPtIlrcoULnZbGIACxS','test@test.com','user','2026-03-18 01:19:11.584','2026-03-18 01:19:11.584','Test',''),(24,'report_test_1773767995','$2a$10$sU0a.MXpCdRmUij838//mOPpFIKi4sT72gFk5NVuR1GtpZjFnYlai','report_test_1773767995@test.com','user','2026-03-18 01:19:55.789','2026-03-18 01:19:55.789','报表测试用户',''),(25,'report_test_1773768004','$2a$10$A1ezS6BPMXwP06YNTm8syeBl0BtH3UXSp2nznYIPGYzyOgzwMUDnC','report_test_1773768004@test.com','user','2026-03-18 01:20:04.527','2026-03-18 01:20:04.527','报表测试用户','');
/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Dumping events for database 'bridge_detection'
--

--
-- Dumping routines for database 'bridge_detection'
--
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-03-18 11:57:56
