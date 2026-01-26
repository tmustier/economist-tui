# economist-tui demo recording
# Run with: vhs demo.tape

Output demo.gif
Set Width 2000
Set Height 1800
Set FontSize 28
Set Theme "Builtin Dark"
Set Padding 40
Set Framerate 60

# Start demo in The Americas section
Type "economist demo the-americas"
Sleep 500ms
Enter
Sleep 2s

# Navigate down to Bay of Pigs article
Down
Sleep 1s

# Open the article
Enter
Sleep 2.5s

# Go back to browse
Escape
Sleep 1.5s

# Tab to previous section (Science & Technology)
Shift+Tab
Sleep 1.5s

# Type to jump to "The Human Use of Human Beings" article
Type "12"
Sleep 1.5s

# Open the article
Enter
Sleep 5s

# Toggle columns
Type "c"
Sleep 2.5s

# Quit
Type "q"
Sleep 500ms
