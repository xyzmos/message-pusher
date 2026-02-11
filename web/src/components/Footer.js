import React, { useEffect, useState } from 'react';

import { Container, Segment } from 'semantic-ui-react';

const Footer = () => {
  const [Footer, setFooter] = useState('');
  useEffect(() => {
    let savedFooter = localStorage.getItem('footer_html');
    if (!savedFooter) savedFooter = '';
    setFooter(savedFooter);
  });

  return (
    <Segment vertical basic>
      <Container textAlign='center'>
        {Footer === '' ? (
          <div className='custom-footer'>
              消息推送服务 {process.env.REACT_APP_VERSION}{' '}
          </div>
        ) : (
          <div
            className='custom-footer'
            dangerouslySetInnerHTML={{ __html: Footer }}
          ></div>
        )}
      </Container>
    </Segment>
  );
};

export default Footer;
