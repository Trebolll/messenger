package ru.lambdahub.common.security;

import org.springframework.boot.autoconfigure.AutoConfiguration;
import org.springframework.boot.autoconfigure.condition.ConditionalOnClass;
import org.springframework.boot.autoconfigure.condition.ConditionalOnMissingBean;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.boot.autoconfigure.condition.ConditionalOnWebApplication;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.boot.security.autoconfigure.web.servlet.ServletWebSecurityAutoConfiguration;
import org.springframework.context.annotation.Bean;
import org.springframework.http.HttpMethod;
import org.springframework.security.config.Customizer;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configurers.AbstractHttpConfigurer;
import org.springframework.security.config.http.SessionCreationPolicy;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.security.web.SecurityFilterChain;
import org.springframework.security.web.authentication.UsernamePasswordAuthenticationFilter;

@AutoConfiguration(before = ServletWebSecurityAutoConfiguration.class)
@ConditionalOnWebApplication(type = ConditionalOnWebApplication.Type.SERVLET)
@ConditionalOnClass(SecurityFilterChain.class)
@ConditionalOnProperty(prefix = "lambdahub.security", name = "enabled", havingValue = "true", matchIfMissing = true)
@EnableConfigurationProperties(LambdahubSecurityProperties.class)
public class LambdahubSecurityAutoConfiguration {

    @Bean
    @ConditionalOnMissingBean(JwtService.class)
    public JwtService jwtService(LambdahubSecurityProperties properties) {
        return new JwtService(properties.getJwtSecret(), properties.getJwtTtlSeconds());
    }

    @Bean
    @ConditionalOnMissingBean(PasswordEncoder.class)
    public PasswordEncoder passwordEncoder() {
        return new BCryptPasswordEncoder();
    }

    @Bean
    @ConditionalOnMissingBean
    public JwtOrHeaderAuthenticationFilter jwtOrHeaderAuthenticationFilter(JwtService jwtService) {
        return new JwtOrHeaderAuthenticationFilter(jwtService);
    }

    @Bean
    @ConditionalOnMissingBean(SecurityFilterChain.class)
    public SecurityFilterChain lambdahubSecurityFilterChain(HttpSecurity http,
                                                            JwtOrHeaderAuthenticationFilter authFilter,
                                                            LambdahubSecurityProperties properties) throws Exception {
        http
                .csrf(AbstractHttpConfigurer::disable)
                .cors(Customizer.withDefaults())
                .sessionManagement(sm -> sm.sessionCreationPolicy(SessionCreationPolicy.STATELESS))
                .authorizeHttpRequests(auth -> {
                    auth.requestMatchers(HttpMethod.OPTIONS, "/**").permitAll();
                    for (String prefix : properties.getPublicPathPrefixes()) {
                        String pattern = prefix.endsWith("/") ? prefix + "**" : prefix + "/**";
                        if (prefix.contains("*")) {
                            auth.requestMatchers(prefix).permitAll();
                        } else if (prefix.endsWith("/")) {
                            auth.requestMatchers(pattern).permitAll();
                        } else {
                            auth.requestMatchers(prefix, pattern).permitAll();
                        }
                    }
                    auth.anyRequest().authenticated();
                })
                .addFilterBefore(authFilter, UsernamePasswordAuthenticationFilter.class);
        return http.build();
    }
}
