package ru.lambdahub.user.db;

import io.vavr.collection.List;
import io.vavr.control.Option;
import java.util.UUID;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

public interface ProfileRepository extends JpaRepository<Profile, UUID> {

    Option<Profile> findByUsernameIgnoreCase(String username);

    @Query("select p from Profile p where lower(p.username) like lower(concat('%', :q, '%')) or lower(p.displayName) like lower(concat('%', :q, '%'))")
    List<Profile> search(@Param("q") String q);

    default Option<Profile> findOne(UUID id) {
        return Option.ofOptional(findById(id));
    }
}
