CREATE TABLE IF NOT EXISTS task_audits (
  id uuid NOT NULL,
  task_id uuid NOT NULL,
  user_id uuid NOT NULL,
  reason text NOT NULL,
  is_refund_requested boolean NOT NULL DEFAULT false,
  admin_feedback text,
  status varchar(20) NOT NULL DEFAULT 'PENDING',
  created_at timestamp DEFAULT NULL,
  updated_at timestamp DEFAULT NULL,
  PRIMARY KEY (id),
  CONSTRAINT fk_task_audits_task FOREIGN KEY (task_id) REFERENCES tasks (id) ON DELETE CASCADE,
  CONSTRAINT fk_task_audits_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);