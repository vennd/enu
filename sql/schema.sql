-- MySQL dump 10.13  Distrib 5.6.24, for Win64 (x86_64)
--
-- Host: 127.0.0.1    Database: vennd
-- ------------------------------------------------------
-- Server version	5.6.25

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
-- Table structure for table `activations`
--

DROP TABLE IF EXISTS `activations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `activations` (
  `rowid` bigint(20) NOT NULL AUTO_INCREMENT,
  `activationId` varchar(45) NOT NULL,
  `blockchainId` varchar(50) NOT NULL,
  `accessKey` varchar(64) NOT NULL,
  `addressToActivate` varchar(200) NOT NULL,
  `amount` bigint(10) NOT NULL,
  PRIMARY KEY (`rowid`)
) ENGINE=InnoDB AUTO_INCREMENT=1529 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `addresses`
--

DROP TABLE IF EXISTS `addresses`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `addresses` (
  `rowId` bigint(20) NOT NULL AUTO_INCREMENT,
  `accessKey` varchar(64) DEFAULT NULL,
  `sourceAddress` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`rowId`)
) ENGINE=InnoDB AUTO_INCREMENT=357 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `addressmaps`
--

DROP TABLE IF EXISTS `addressmaps`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `addressmaps` (
  `rowid` bigint(20) NOT NULL AUTO_INCREMENT,
  `counterpartyPaymentAddress` varchar(200) DEFAULT NULL,
  `nativePaymentAddress` varchar(200) DEFAULT NULL,
  `externalAddress` varchar(200) DEFAULT NULL,
  `counterpartyAddress` varchar(200) DEFAULT NULL,
  `counterpartyAssetName` varchar(200) DEFAULT NULL,
  `nativeAssetName` varchar(200) DEFAULT NULL,
  `UDF1` varchar(200) DEFAULT NULL,
  `UDF2` varchar(200) DEFAULT NULL,
  `UDF3` varchar(200) DEFAULT NULL,
  `UDF4` varchar(200) DEFAULT NULL,
  `UDF5` varchar(200) DEFAULT NULL,
  PRIMARY KEY (`rowid`),
  UNIQUE KEY `addressMaps1` (`counterpartyPaymentAddress`),
  UNIQUE KEY `addressMaps2` (`nativePaymentAddress`),
  UNIQUE KEY `addressMaps3` (`externalAddress`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `assets`
--

DROP TABLE IF EXISTS `assets`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `assets` (
  `rowid` bigint(20) NOT NULL AUTO_INCREMENT,
  `accessKey` varchar(64) DEFAULT NULL,
  `assetId` varchar(200) DEFAULT NULL,
  `sourceAddress` varchar(200) DEFAULT NULL,
  `distributionAddress` varchar(200) DEFAULT NULL,
  `asset` varchar(200) DEFAULT NULL,
  `blockchainId` varchar(50) DEFAULT NULL,
  `description` varchar(200) DEFAULT NULL,
  `quantity` bigint(20) DEFAULT NULL,
  `divisible` tinyint(1) DEFAULT NULL,
  `status` varchar(200) DEFAULT NULL,
  `broadcastTxId` varchar(200) DEFAULT NULL,
  `blockchainStatus` varchar(10) DEFAULT NULL,
  `errorCode` bigint(20) DEFAULT NULL,
  `errorDescription` varchar(512) DEFAULT NULL,
  `requestId` varchar(200) DEFAULT NULL,
  `retryCount` tinyint(4) DEFAULT NULL,
  `signedRawTx` text,
  `issuer` varchar(200) DEFAULT NULL,
  PRIMARY KEY (`rowid`),
  KEY `assets1` (`assetId`)
) ENGINE=InnoDB AUTO_INCREMENT=255 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `audit`
--

DROP TABLE IF EXISTS `audit`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `audit` (
  `rowid` bigint(20) NOT NULL AUTO_INCREMENT,
  `blockId` varchar(200) DEFAULT NULL,
  `txid` varchar(200) DEFAULT NULL,
  `description` varchar(200) DEFAULT NULL,
  PRIMARY KEY (`rowid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `blockchains`
--

DROP TABLE IF EXISTS `blockchains`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `blockchains` (
  `rowId` bigint(20) NOT NULL AUTO_INCREMENT,
  `blockchainId` varchar(50) DEFAULT NULL,
  PRIMARY KEY (`rowId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `blocks`
--

DROP TABLE IF EXISTS `blocks`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `blocks` (
  `rowid` bigint(20) NOT NULL AUTO_INCREMENT,
  `blockId` bigint(20) DEFAULT NULL,
  `status` varchar(100) DEFAULT NULL,
  `duration` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`rowid`),
  UNIQUE KEY `blocks1` (`blockId`)
) ENGINE=InnoDB AUTO_INCREMENT=331 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `credits`
--

DROP TABLE IF EXISTS `credits`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `credits` (
  `rowid` bigint(20) NOT NULL AUTO_INCREMENT,
  `blockIdSource` bigint(20) DEFAULT NULL,
  `txid` varchar(200) DEFAULT NULL,
  `sourceAddress` varchar(200) DEFAULT NULL,
  `destinationAddress` varchar(200) DEFAULT NULL,
  `inAsset` varchar(200) DEFAULT NULL,
  `inAmount` bigint(20) DEFAULT NULL,
  `outAsset` varchar(200) DEFAULT NULL,
  `outAmount` bigint(20) DEFAULT NULL,
  `status` varchar(200) DEFAULT NULL,
  PRIMARY KEY (`rowid`),
  KEY `credits1` (`blockIdSource`),
  KEY `credits2` (`txid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `debits`
--

DROP TABLE IF EXISTS `debits`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `debits` (
  `rowid` bigint(20) NOT NULL AUTO_INCREMENT,
  `blockIdSource` bigint(20) DEFAULT NULL,
  `txid` varchar(200) DEFAULT NULL,
  `sourceAddress` varchar(200) DEFAULT NULL,
  `destinationAddress` varchar(200) DEFAULT NULL,
  `inAsset` varchar(200) DEFAULT NULL,
  `inAmount` bigint(20) DEFAULT NULL,
  `outAsset` varchar(200) DEFAULT NULL,
  `outAmount` bigint(20) DEFAULT NULL,
  `status` varchar(200) DEFAULT NULL,
  `lastUpdatedBlockId` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`rowid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dividends`
--

DROP TABLE IF EXISTS `dividends`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dividends` (
  `rowid` bigint(20) NOT NULL AUTO_INCREMENT,
  `accessKey` varchar(64) DEFAULT NULL,
  `dividendId` varchar(200) DEFAULT NULL,
  `sourceAddress` varchar(200) DEFAULT NULL,
  `asset` varchar(200) DEFAULT NULL,
  `dividendAsset` varchar(200) DEFAULT NULL,
  `quantityPerUnit` bigint(20) DEFAULT NULL,
  `status` varchar(200) DEFAULT NULL,
  `broadcastTxId` varchar(200) DEFAULT NULL,
  `blockchainStatus` varchar(10) DEFAULT NULL,
  `errorCode` bigint(20) DEFAULT NULL,
  `errorDescription` varchar(512) DEFAULT NULL,
  `retryCount` tinyint(4) DEFAULT NULL,
  `signedRawTx` text,
  PRIMARY KEY (`rowid`),
  KEY `dividends1` (`dividendId`)
) ENGINE=InnoDB AUTO_INCREMENT=145 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fees`
--

DROP TABLE IF EXISTS `fees`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fees` (
  `rowid` bigint(20) NOT NULL AUTO_INCREMENT,
  `blockId` varchar(200) DEFAULT NULL,
  `txid` varchar(200) DEFAULT NULL,
  `feeAsset` varchar(200) DEFAULT NULL,
  `feeAmount` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`rowid`),
  KEY `fees1` (`blockId`,`txid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inputaddresses`
--

DROP TABLE IF EXISTS `inputaddresses`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `inputaddresses` (
  `rowid` bigint(20) NOT NULL AUTO_INCREMENT,
  `txid` varchar(200) DEFAULT NULL,
  `address` varchar(200) DEFAULT NULL,
  PRIMARY KEY (`rowid`),
  KEY `inputAddresses1` (`txid`),
  KEY `inputAddresses2` (`address`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `outputaddresses`
--

DROP TABLE IF EXISTS `outputaddresses`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `outputaddresses` (
  `rowid` bigint(20) NOT NULL AUTO_INCREMENT,
  `txid` varchar(200) DEFAULT NULL,
  `address` varchar(200) DEFAULT NULL,
  PRIMARY KEY (`rowid`),
  KEY `outputAddresses1` (`txid`),
  KEY `outputAddresses2` (`address`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `payments`
--

DROP TABLE IF EXISTS `payments`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `payments` (
  `rowid` bigint(20) NOT NULL AUTO_INCREMENT,
  `accessKey` varchar(64) DEFAULT NULL,
  `blockId` bigint(20) DEFAULT NULL,
  `blockchainId` varchar(50) DEFAULT NULL,
  `sourceTxid` varchar(200) DEFAULT NULL,
  `sourceAddress` varchar(200) DEFAULT NULL,
  `destinationAddress` varchar(200) DEFAULT NULL,
  `outAsset` varchar(200) DEFAULT NULL,
  `issuer` varchar(200) DEFAULT NULL,
  `outAmount` bigint(20) DEFAULT NULL,
  `status` varchar(200) DEFAULT NULL,
  `lastUpdatedBlockId` bigint(20) DEFAULT NULL,
  `txFee` bigint(20) DEFAULT NULL,
  `broadcastTxId` varchar(200) DEFAULT NULL,
  `blockchainStatus` varchar(10) DEFAULT NULL,
  `errorCode` bigint(20) DEFAULT NULL,
  `errorDescription` varchar(512) DEFAULT NULL,
  `paymentTag` varchar(512) DEFAULT NULL,
  `retryCount` tinyint(4) DEFAULT NULL,
  `signedRawTx` text,
  PRIMARY KEY (`rowid`),
  KEY `payments1` (`blockId`)
) ENGINE=InnoDB AUTO_INCREMENT=1742 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `registry`
--

DROP TABLE IF EXISTS `registry`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `registry` (
  `rowid` bigint(20) NOT NULL AUTO_INCREMENT,
  `blockId` varchar(200) DEFAULT NULL,
  `ownerAddress` varchar(200) DEFAULT NULL,
  `asset` varchar(200) DEFAULT NULL,
  `quantity` bigint(20) DEFAULT NULL,
  `status` varchar(200) DEFAULT NULL,
  PRIMARY KEY (`rowid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `transactions`
--

DROP TABLE IF EXISTS `transactions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `transactions` (
  `rowid` bigint(20) NOT NULL AUTO_INCREMENT,
  `blockId` bigint(20) DEFAULT NULL,
  `txid` varchar(200) DEFAULT NULL,
  PRIMARY KEY (`rowid`),
  KEY `transactions1` (`blockId`),
  KEY `transactions2` (`txid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `trustassets`
--

DROP TABLE IF EXISTS `trustassets`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `trustassets` (
  `rowid` bigint(20) NOT NULL AUTO_INCREMENT,
  `activationId` varchar(45) DEFAULT NULL,
  `blockchainId` varchar(50) DEFAULT NULL,
  `accessKey` varchar(64) DEFAULT NULL,
  `asset` varchar(200) DEFAULT NULL,
  `issuer` varchar(200) DEFAULT NULL,
  `trustAmount` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`rowid`)
) ENGINE=InnoDB AUTO_INCREMENT=74 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `userkeys`
--

DROP TABLE IF EXISTS `userkeys`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `userkeys` (
  `rowId` bigint(20) NOT NULL AUTO_INCREMENT,
  `userId` bigint(20) DEFAULT NULL,
  `parentAccessKey` varchar(64) DEFAULT NULL,
  `accessKey` varchar(64) DEFAULT NULL,
  `secret` varchar(64) DEFAULT NULL,
  `nonce` bigint(20) DEFAULT NULL,
  `assetId` varchar(100) DEFAULT NULL,
  `blockchainId` varchar(100) DEFAULT NULL,
  `status` varchar(10) DEFAULT NULL,
  PRIMARY KEY (`rowId`)
) ENGINE=InnoDB AUTO_INCREMENT=337 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2015-12-10 19:17:51
