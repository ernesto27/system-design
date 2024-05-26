CREATE TABLE `user` (
  `id` INT AUTO_INCREMENT,
  `email` VARCHAR(255) NOT NULL,
  `password` VARCHAR(255) NOT NULL,
  PRIMARY KEY (`id`)
);


CREATE TABLE `file` (
  `id` INT AUTO_INCREMENT,
  `name` VARCHAR(255) NOT NULL,
  `hash` VARCHAR(255) NOT NULL,
  `fileSize` BIGINT NOT NULL,
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  `owner_id` INT,
  PRIMARY KEY (`id`),
  FOREIGN KEY (`owner_id`) REFERENCES `user`(`id`)
);