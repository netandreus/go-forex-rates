-- MySQL dump 10.13  Distrib 5.7.32, for osx10.15 (x86_64)
--
-- Host: localhost    Database: go_forex_rates
-- ------------------------------------------------------
-- Server version	8.0.25

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `currency_rate`
--

DROP TABLE IF EXISTS `currency_rate`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `currency_rate` (
  `id` int NOT NULL AUTO_INCREMENT,
  `base_currency` char(3) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL COMMENT 'Base currency',
  `quoted_currency` char(3) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL COMMENT 'Quoted currency',
  `value` double NOT NULL,
  `rate_date` date NOT NULL,
  `request_time` datetime NOT NULL,
  `provider_generated_time` datetime NOT NULL,
  `provider` enum('fixer','emirates') CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL DEFAULT 'fixer',
  `endpoint` enum('historical','latest') CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL DEFAULT 'historical',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_currency_rate_quoted_base_date_endpoint` (`quoted_currency`,`base_currency`,`rate_date`,`provider`,`endpoint`)
) ENGINE=InnoDB AUTO_INCREMENT=292583 DEFAULT CHARSET=utf8mb3 COLLATE=utf8_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
CREATE INDEX currency_rate_endpoint_provider_index ON currency_rate (endpoint, provider);
CREATE INDEX currency_rate_endpoint_provider_rate_date_index ON currency_rate (endpoint, provider, rate_date);
--
-- Dumping data for table `currency_rate`
--

LOCK TABLES `currency_rate` WRITE;
/*!40000 ALTER TABLE `currency_rate` DISABLE KEYS */;
/*!40000 ALTER TABLE `currency_rate` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2021-08-04 18:15:54
