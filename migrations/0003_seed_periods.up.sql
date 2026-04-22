INSERT INTO periods (code, title, rrule_template) VALUES
('daily', 'Ежедневно', 'FREQ=DAILY'),
('weekdays', 'По будням', 'FREQ=DAILY;BYDAY=MO,TU,WE,TH,FR'),
('weekly', 'Еженедельно', 'FREQ=WEEKLY'),
('monthly_same_day', 'Ежемесячно (в тот же день)', 'FREQ=MONTHLY'),
('even_days', 'По чётным дням', 'FREQ=DAILY;BYMONTHDAY=2,4,6,8,10,12,14,16,18,20,22,24,26,28,30'),
('odd_days', 'По нечётным дням', 'FREQ=DAILY;BYMONTHDAY=1,3,5,7,9,11,13,15,17,19,21,23,25,27,29,31');
