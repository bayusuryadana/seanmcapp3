import Typography, { TypographyProps } from '@mui/material/Typography';

interface TitleProps extends TypographyProps {}

export const Title = ({ children, sx, ...props }: TitleProps) => {
  return (
    <Typography
      component="h2"
      variant="h6"
      color="primary"
      gutterBottom
      sx={{ display: 'inline', ...sx }}
      {...props}
    >
      {children}
    </Typography>
  );
};