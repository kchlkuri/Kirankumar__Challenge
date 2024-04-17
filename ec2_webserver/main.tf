provider "aws" {
  region = "us-east-1"
}

data "template_cloudinit_config" "init" {
  gzip          = true
  base64_encode = true

  part {
    content_type = "text/x-shellscript"
    content      = "${file("scripts/install.sh")}"
 }

}


resource "aws_launch_configuration" "demo_lc" {
  name_prefix   = "${var.application}-lc"
  image_id             = var.ami_id
  key_name     = var.keypair
  security_groups      = [aws_security_group.web_asg.id]
  user_data_base64     = data.template_cloudinit_config.init.rendered
  enable_monitoring    = var.enable_monitor
  ebs_optimized        = var.enable_ebs_optimization
  instance_type        = var.instance_type
  iam_instance_profile = var.instance_profile

  root_block_device {
    volume_type           = var.root_ebs_type
    volume_size           = var.root_ebs_size
    delete_on_termination = var.root_ebs_del_on_term
    
  }

  associate_public_ip_address = var.associate_public_ip_address
  
}

resource "aws_autoscaling_group" "web_asg" {
  name                      = "${var.application}-asg"
  max_size                  = var.asg_max_cap
  min_size                  = var.asg_min_cap
  desired_capacity          = var.asg_desired_cap
  health_check_grace_period = var.asg_health_check_grace_period
  health_check_type         = var.asg_health_check_type
  force_delete              = var.asg_force_delete
  termination_policies      = var.asg_termination_policies
  suspended_processes       = var.asg_suspended_processes
  launch_configuration      = aws_launch_configuration.demo_lc.name
  vpc_zone_identifier       = var.alb_subnets
  default_cooldown          = var.asg_default_cooldown
  enabled_metrics           = var.asg_enabled_metrics
  metrics_granularity       = var.asg_metrics_granularity
  protect_from_scale_in     = var.asg_protect_from_scale_in
  target_group_arns         = ["${aws_lb_target_group.webtg1.arn}"]

}

resource "aws_cloudwatch_metric_alarm" "demo-cpu-alarm" {
  alarm_name          = "demo-${var.application}-cpu-alarm"
  alarm_description   = "demo-${var.application}-cpu-alarm"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = "2"
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = "120"
  statistic           = "Average"
  threshold           = "80"

  dimensions = {
    "AutoScalingGroupapplication" = aws_autoscaling_group.web_asg.name
  }

  actions_enabled = true
}

resource "aws_security_group" "web_asg" {
  name        = "${var.application}-security-group"
  description = "Allow demo traffic"
  vpc_id      = var.vpc_id

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_cloudwatch_metric_alarm" "target-healthy-count" {
  alarm_name          = "${var.application}-tg1-Healthy-Count"
  comparison_operator = "LessThanOrEqualToThreshold"
  evaluation_periods  = "1"
  metric_name         = "HealthyHostCount"
  namespace           = "AWS/ApplicationELB"
  period              = "60"
  statistic           = "Average"
  threshold           = "0"

  dimensions = {
    LoadBalancer = "${aws_lb.demo.arn_suffix}"
    TargetGroup  = "${aws_lb_target_group.webtg1.arn_suffix}"
  }
}

resource "aws_security_group" "elb" {
  name        = "${var.application}-alb-sg"
  description = "ALB SG"

  vpc_id = var.vpc_id

  # HTTP access from anywhere
   ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  } 

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  } 

  # outbound internet access
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

}


resource "aws_lb" "demo" {
  name               = "${var.application}-demo"
  internal           = false
  load_balancer_type = "application"
  security_groups    = ["${aws_security_group.elb.id}"]
  subnets            =   var.alb_subnets

}

resource "aws_lb_listener" "web_tg1" {
  load_balancer_arn = "${aws_lb.demo.arn}"
  port              = "80"
  protocol          = "HTTP"
  
  default_action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.webtg1.arn}"
  }
}

resource "aws_lb_listener" "web_tg2" {
  load_balancer_arn = "${aws_lb.demo.arn}"
  port              = "443"
  protocol          = "HTTPS"
  
  default_action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.webtg1.arn}"
  }
}



resource "aws_lb_target_group" "webtg1" {
  name     = "${var.application}-tg1"
  port     = 443
  protocol = "HTTPS"
  vpc_id   = var.vpc_id

health_check {
                path = "/index.html"
                port = "443"
                protocol = "HTTPS"
                healthy_threshold = 2
                unhealthy_threshold = 2
                interval = 5
                timeout = 4
                matcher = "200-308"
        }
}