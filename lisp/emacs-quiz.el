;; filtra los comandos inútiles
(defun zr-comando-util-p (cmd)
  (and (symbolp cmd)
       (commandp cmd)
       (let ((name (symbol-name cmd)))
	 (and (not (string= name "self-insert-command"))
	      (not (string-match-p "-prefix\\'" name))
	      (not (string-match-p "\\`\\(universal\\|digit\\|negative\\)-argument" name))
	      (not (string-match-p "\\`mouse-" name))
	      (not (string-match-p "\\`menu-" name))
	      (not (string-match-p "\\`scroll-" name))))))

;; obtiene un atajo de teclado al azar
(defun zr-obtener-binding-azar ()
  "Devuelve (\"C-x C-f\" . \"find-file\") de forma segura."
  (let (bindings)
    (map-keymap
     (lambda (key binding)
       (when (and (symbolp binding)
                  (zr-comando-util-p binding))
         (when (or (integerp key) (vectorp key))
           (let ((key-desc (key-description (if (integerp key) (vector key) key))))
             (unless (or (string-match-p "<key>" key-desc)
                         (string-match-p "\\`[0-9]" key-desc)
                         (string-empty-p key-desc)
                         (string-match-p "mouse\\|menu\\|wheel\\|xterm" key-desc))
               (push (cons key-desc (symbol-name binding)) bindings))))))
     (current-global-map))

    (when bindings
      (seq-random-elt (delete-dups bindings)))))
